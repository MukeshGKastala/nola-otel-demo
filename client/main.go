package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type CreateCalculationRequest struct {
	Expression string `json:"expression"`
	Student    string `json:"student"`
}

type CreateCalculationResponse struct {
	Id uuid.UUID `json:"id"`
}

func main() {
	ctx := context.Background()

	//Register global trace provider.
	tp, err := otelcommon.InitTracer(ctx, otelcommon.Config{
		ServiceName:    "client",
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

	if err := func() error {
		url := "http://localhost:80/calculations"
		calc := CreateCalculationRequest{
			Expression: "8 + 12",
			Student:    "go client",
		}
		b, err := json.Marshal(calc)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return err
		}

		client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		b, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var resp CreateCalculationResponse
		if err := json.Unmarshal(b, &resp); err != nil {
			return err
		}

		log.Println("Calculation Id:", resp.Id)
		return nil
	}(); err != nil {
		log.Fatal(err)
	}
}
