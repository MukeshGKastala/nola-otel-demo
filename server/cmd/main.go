package main

import (
	"context"
	"log"
	"net/http"
	"os"

	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	api "github.com/MukeshGKastala/nola-otel-demo/server/api/calculator/v1"
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

	_, err = postgres.NewAndMigrate(ctx, postgres.Config{
		Host:         os.Getenv("POSTGRES_HOST"),
		User:         os.Getenv("POSTGRES_USER"),
		Password:     os.Getenv("POSTGRES_PASSWORD"),
		DatabaseName: os.Getenv("POSTGRES_DB"),
	})
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Handler: api.MakeHTTPHandler(service.NewService()),
	}

	log.Fatal(server.ListenAndServe())
}
