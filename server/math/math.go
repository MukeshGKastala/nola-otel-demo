package math

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"time"

	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	"github.com/MukeshGKastala/nola-otel-demo/server/store/postgres"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/udhos/opentelemetry-trace-sqs/otelsqs"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Store interface {
	UpdateCalculation(context.Context, postgres.UpdateCalculationParams) (postgres.Calculation, error)
}

type Calculation struct {
	ID         uuid.UUID `json:"id"`
	Student    string    `json:"student"`
	Expression string    `json:"expression"`
}

type Config struct {
	SQSRegion         string
	SQSBaseEndpoint   string
	SQSReadQueueName  string
	SQSWriteQueueName string
}

type handler struct {
	client        *sqs.Client
	readQueueUrl  string
	writeQueueUrl string
	store         Store
}

func New(ctx context.Context, cfg Config, store Store) (*handler, error) {
	c := sqs.New(sqs.Options{
		Region:       cfg.SQSRegion,
		BaseEndpoint: aws.String(cfg.SQSBaseEndpoint),
	})

	resp, err := c.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(cfg.SQSReadQueueName),
	})
	if err != nil {
		return nil, err
	}

	readQueueUrl := *resp.QueueUrl

	resp, err = c.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(cfg.SQSWriteQueueName),
	})
	if err != nil {
		return nil, err
	}

	writeQueueUrl := *resp.QueueUrl

	h := &handler{
		client:        c,
		readQueueUrl:  readQueueUrl,
		writeQueueUrl: writeQueueUrl,
		store:         store,
	}

	go func() {
		if err := h.receiveMessages(ctx); err != nil {
			log.Printf("unable to receive queue messages: %v", err)
		}
	}()

	return h, nil
}

func (h *handler) receiveMessages(ctx context.Context) error {
	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{"b3"},
		QueueUrl:              aws.String(h.readQueueUrl),
		VisibilityTimeout:     60,
		WaitTimeSeconds:       10,
	}

	for {
		resp, err := h.client.ReceiveMessage(ctx, gMInput)
		if err != nil {
			return err
		}

		for _, msg := range resp.Messages {
			ctx := otelsqs.NewCarrier().Extract(msg.MessageAttributes)
			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(
					semconv.MessagingSystemKey.String("elasticmq"),
					semconv.MessagingDestinationName(path.Base(h.readQueueUrl)),
					semconv.MessagingMessageID(*msg.MessageId),
				),
			}
			ctx, span := otelcommon.Tracer().Start(ctx, fmt.Sprintf("%s process", path.Base(h.readQueueUrl)), opts...)

			var rslt struct {
				Id     uuid.UUID `json:"id"`
				Result float64   `json:"result"`
			}
			if err := json.Unmarshal([]byte(*msg.Body), &rslt); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return err
			}

			if _, err := h.store.UpdateCalculation(ctx, postgres.UpdateCalculationParams{
				ID: rslt.Id,
				Result: pgtype.Float8{
					Float64: rslt.Result,
					Valid:   true,
				},
				Completed: pgtype.Timestamptz{
					Time:  time.Now(),
					Valid: true,
				},
			}); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return err
			}

			if _, err := h.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(h.readQueueUrl),
				ReceiptHandle: msg.ReceiptHandle,
			}); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return err
			}

			span.End()
		}
	}
}

func (h *handler) Calculate(ctx context.Context, calc Calculation) error {
	b, err := json.Marshal(calc)
	if err != nil {
		return err
	}

	sMInput := &sqs.SendMessageInput{
		MessageAttributes: map[string]types.MessageAttributeValue{},
		MessageBody:       aws.String(string(b)),
		QueueUrl:          &h.writeQueueUrl,
	}

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("elasticmq"),
			semconv.MessagingDestinationName(path.Base(h.writeQueueUrl)),
		),
	}
	ctx, span := otelcommon.Tracer().Start(ctx, fmt.Sprintf("%s send", path.Base(h.writeQueueUrl)), opts...)
	otelsqs.NewCarrier().Inject(ctx, sMInput.MessageAttributes)
	defer span.End()

	resp, err := h.client.SendMessage(ctx, sMInput)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(semconv.MessagingMessageID(*resp.MessageId))

	return nil
}
