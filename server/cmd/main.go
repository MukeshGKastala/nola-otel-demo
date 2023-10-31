package main

import (
	"context"
	"log"
	"net/http"
	"os"

	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	api "github.com/MukeshGKastala/nola-otel-demo/server/api/calculator/v1"
	"github.com/MukeshGKastala/nola-otel-demo/server/math"
	"github.com/MukeshGKastala/nola-otel-demo/server/service"
	"github.com/MukeshGKastala/nola-otel-demo/server/store/postgres"
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

	conn, err := postgres.ConnectAndMigrate(ctx, postgres.Config{
		Host:         os.Getenv("POSTGRES_HOST"),
		User:         os.Getenv("POSTGRES_USER"),
		Password:     os.Getenv("POSTGRES_PASSWORD"),
		DatabaseName: os.Getenv("POSTGRES_DB"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	store := postgres.New(conn)

	calculator, err := math.New(ctx, math.Config{
		SQSRegion:         os.Getenv("SQS_REGION"),
		SQSBaseEndpoint:   os.Getenv("SQS_BASE_ENDPOINT"),
		SQSReadQueueName:  os.Getenv("SQS_READ_QUEUE_NAME"),
		SQSWriteQueueName: os.Getenv("SQS_WRITE_QUEUE_NAME"),
	}, store)
	if err != nil {
		log.Fatal(err)
	}

	svc := service.NewService(store, calculator)

	server := &http.Server{
		Handler: api.MakeHTTPHandler(svc),
	}

	log.Fatal(server.ListenAndServe())
}
