package main

import (
	"context"
	"encoding/json"
	"errors"
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

func main() {

	ctx := context.Background()

	// Register global trace provider.
	tp, err := otelcommon.InitTracer(ctx, otelcommon.Config{
		ServiceName:    "server",
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
		QueueName: aws.String("math-queue"),
	})
	if err != nil {
		log.Fatal(err)
	}

	requestQueueUrl := *resp.QueueUrl

	resp, err = c.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String("math-result-queue"),
	})
	if err != nil {
		log.Fatal(err)
	}

	responseQueueUrl := *resp.QueueUrl

	receive(context.Background(), requestQueueUrl, c, responseQueueUrl)
}

func receive(ctx context.Context, requestQueueUrl string, client *sqs.Client, responseQueueUrl string) error {
	gMInput := &sqs.ReceiveMessageInput{
		MessageAttributeNames: []string{"b3"},
		QueueUrl:              aws.String(requestQueueUrl),
		VisibilityTimeout:     60,
		WaitTimeSeconds:       10,
	}

	for {
		resp, err := client.ReceiveMessage(ctx, gMInput)
		if err != nil {
			return err
		}

		if len(resp.Messages) == 0 {
			continue
		}

		if len(resp.Messages) != 1 {
			return errors.New("unexpected number of requests")
		}

		msg := resp.Messages[0]

		ctx := otelsqs.NewCarrier().Extract(msg.MessageAttributes)
		ctx, span := otelcommon.Tracer().Start(ctx, fmt.Sprintf("%s process", path.Base(requestQueueUrl)), []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindConsumer),
		}...)

		var r struct {
			ID         uuid.UUID `json:"id"`
			Student    string    `json:"student"`
			Expression string    `json:"expression"`
		}
		if err := json.Unmarshal([]byte(*msg.Body), &r); err != nil {
			return err
		}

		if r.Student == "lazy" {
			time.Sleep(15 * time.Millisecond)
		}

		eval := goval.NewEvaluator()
		result, err := eval.Evaluate(r.Expression, nil, nil)
		if err != nil {
			return err
		}

		var ret float64
		if intResult, ok := result.(int); ok {
			ret = float64(intResult)
		} else if floatResult, ok := result.(float64); ok {
			ret = floatResult
		}

		if err := send(ctx, responseQueueUrl, client, r.ID, ret); err != nil {
			return err
		}

		if _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(requestQueueUrl),
			ReceiptHandle: msg.ReceiptHandle,
		}); err != nil {
			return err
		}

		span.End()
	}
}

func send(ctx context.Context, responseQueueUrl string, client *sqs.Client, id uuid.UUID, result float64) error {
	ctx, span := otelcommon.Tracer().Start(ctx, fmt.Sprintf("%s send", path.Base(responseQueueUrl)), []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindProducer),
	}...)
	defer span.End()

	r := struct {
		Id     uuid.UUID `json:"id"`
		Result float64   `json:"result"`
	}{
		Id:     id,
		Result: result,
	}

	b, err := json.Marshal(r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	sMInput := &sqs.SendMessageInput{
		MessageAttributes: map[string]types.MessageAttributeValue{},
		MessageBody:       aws.String(string(b)),
		QueueUrl:          aws.String(responseQueueUrl),
	}

	otelsqs.NewCarrier().Inject(ctx, sMInput.MessageAttributes)

	resp, err := client.SendMessage(ctx, sMInput)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(semconv.MessagingMessageIDKey.String(*resp.MessageId))

	return err
}
