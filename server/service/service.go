package service

import (
	"context"
	"time"

	api "github.com/MukeshGKastala/nola-otel-demo/api/calculator/v1"
	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	"go.opentelemetry.io/otel/attribute"
)

type service struct {
}

var _ api.StrictServerInterface = (*service)(nil)

func NewsSrvice() *service {
	return &service{}
}

func (s *service) CreateCalculation(ctx context.Context, request api.CreateCalculationRequestObject) (api.CreateCalculationResponseObject, error) {
	_, span := otelcommon.Tracer().Start(ctx, "CreateCalculation Service")
	defer span.End()

	span.SetAttributes(
		attribute.String("expression", request.Body.Expression),
		attribute.String("owner", request.Body.Owner),
	)

	// Imitate work
	time.Sleep(1 * time.Second)

	return api.CreateCalculation200JSONResponse{
		Created:    time.Now(),
		Expression: request.Body.Expression,
		Owner:      request.Body.Owner,
	}, nil
}
