package main

import (
	"context"
	"log"
	"net/http"

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
		Host:         "localhost:5432",
		User:         "admin",
		Password:     "admin",
		DatabaseName: "nola_otel_demo_db",
	})
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Handler: api.MakeHTTPHandler(service.NewsSrvice()),
	}

	log.Fatal(server.ListenAndServe())
}
