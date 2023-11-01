package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/maja42/goval"
	"github.com/udhos/opentelemetry-trace-sqs/otelsqs"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type problem struct {
	ID         uuid.UUID `json:"id"`
	Student    string    `json:"student"`
	Expression string    `json:"expression"`
}

type solution struct {
	ID     uuid.UUID `json:"id"`
	Result float64   `json:"result"`
}

type calculator struct {
	client        *sqs.Client
	readQueueUrl  string
	writeQueueUrl string
}

func (c *calculator) calculate(ctx context.Context) error {
	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{"b3"},
		QueueUrl:              aws.String(c.readQueueUrl),
		VisibilityTimeout:     60,
		WaitTimeSeconds:       10,
	}

	for {
		resp, err := c.client.ReceiveMessage(ctx, gMInput)
		if err != nil {
			return err
		}

		for _, msg := range resp.Messages {
			ctx := otelsqs.NewCarrier().Extract(msg.MessageAttributes)
			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindConsumer),
				trace.WithAttributes(
					semconv.MessagingSystemKey.String("elasticmq"),
					semconv.MessagingDestinationName(path.Base(c.readQueueUrl)),
					semconv.MessagingMessageID(*msg.MessageId),
				),
			}
			ctx, span := otelcommon.Tracer().Start(ctx, fmt.Sprintf("%s process", path.Base(c.readQueueUrl)), opts...)

			var p problem
			if err := json.Unmarshal([]byte(*msg.Body), &p); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return err
			}

			if p.Student == "lazy" {
				time.Sleep(15 * time.Millisecond)
			}

			v, err := goval.NewEvaluator().Evaluate(p.Expression, nil, nil)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return err
			}

			var result float64
			if n, ok := v.(int); ok {
				result = float64(n)
			} else if f, ok := v.(float64); ok {
				result = f
			}

			s := solution{p.ID, result}
			if err := c.enqueueSolution(ctx, s); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				span.End()
				return err
			}

			if _, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(c.readQueueUrl),
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

func (c *calculator) enqueueSolution(ctx context.Context, s solution) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	sMInput := &sqs.SendMessageInput{
		MessageAttributes: map[string]types.MessageAttributeValue{},
		MessageBody:       aws.String(string(b)),
		QueueUrl:          aws.String(c.writeQueueUrl),
	}

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("elasticmq"),
			semconv.MessagingDestinationName(path.Base(c.writeQueueUrl)),
		),
	}
	ctx, span := otelcommon.Tracer().Start(ctx, fmt.Sprintf("%s send", path.Base(c.writeQueueUrl)), opts...)
	otelsqs.NewCarrier().Inject(ctx, sMInput.MessageAttributes)
	defer span.End()

	resp, err := c.client.SendMessage(ctx, sMInput)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(semconv.MessagingMessageIDKey.String(*resp.MessageId))

	return nil
}

func main() {
	ctx := context.Background()

	// Register global trace provider.
	tp, err := otelcommon.InitTracer(ctx, otelcommon.Config{
		ServiceName:    "calculator",
		ServiceVersion: "v0.0.1",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	c := sqs.New(sqs.Options{
		Region:       os.Getenv("SQS_REGION"),
		BaseEndpoint: aws.String(os.Getenv("SQS_BASE_ENDPOINT")),
	})

	resp, err := c.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(os.Getenv("SQS_READ_QUEUE_NAME")),
	})
	if err != nil {
		log.Fatal(err)
	}

	readQueueUrl := *resp.QueueUrl

	resp, err = c.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(os.Getenv("SQS_WRITE_QUEUE_NAME")),
	})
	if err != nil {
		log.Fatal(err)
	}

	writeQueueUrl := *resp.QueueUrl

	calc := calculator{
		client:        c,
		readQueueUrl:  readQueueUrl,
		writeQueueUrl: writeQueueUrl,
	}

	log.Fatal(calc.calculate(ctx))
}
