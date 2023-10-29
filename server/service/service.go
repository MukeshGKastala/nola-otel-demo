package service

import (
	"context"
	"net/http"

	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	api "github.com/MukeshGKastala/nola-otel-demo/server/api/calculator/v1"
	"github.com/MukeshGKastala/nola-otel-demo/server/store/postgres"
	"go.opentelemetry.io/otel/attribute"
)

type service struct {
	store postgres.Querier
}

func NewService(store postgres.Querier) *service {
	return &service{store: store}
}

func (s *service) CreateCalculation(ctx context.Context, request api.CreateCalculationRequestObject) (api.CreateCalculationResponseObject, error) {
	_, span := otelcommon.Tracer().Start(ctx, "CreateCalculation Service")
	defer span.End()

	span.SetAttributes(
		attribute.String("expression", request.Body.Expression),
		attribute.String("student", request.Body.Student),
	)

	calc, err := s.store.CreateCalculation(ctx, postgres.CreateCalculationParams{
		Student:    request.Body.Student,
		Expression: request.Body.Expression,
	})
	if err != nil {
		return api.CreateCalculationdefaultJSONResponse{
			StatusCode: http.StatusInternalServerError,
			Body: api.Error{
				Message: "database write failure",
			},
		}, nil
	}

	return api.CreateCalculation200JSONResponse{
		Id:         calc.ID,
		Created:    calc.Created,
		Expression: calc.Expression,
		Student:    calc.Student,
	}, nil
}
