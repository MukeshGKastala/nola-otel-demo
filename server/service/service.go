package service

import (
	"context"
	"net/http"
	"time"

	otelcommon "github.com/MukeshGKastala/nola-otel-demo/common/otel"
	api "github.com/MukeshGKastala/nola-otel-demo/server/api/calculator/v1"
	"github.com/MukeshGKastala/nola-otel-demo/server/math"
	"github.com/MukeshGKastala/nola-otel-demo/server/store/postgres"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Store interface {
	CreateCalculation(context.Context, postgres.CreateCalculationParams) (uuid.UUID, error)
	GetCalculation(context.Context, uuid.UUID) (postgres.Calculation, error)
}

type Math interface {
	Calculate(context.Context, math.Calculation) error
}

type service struct {
	store Store
	math  Math
}

func NewService(store Store, math Math) *service {
	return &service{store: store, math: math}
}

func (s *service) CreateCalculation(ctx context.Context, request api.CreateCalculationRequestObject) (api.CreateCalculationResponseObject, error) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("expression", request.Body.Expression),
			attribute.String("student", request.Body.Student),
		),
	}
	ctx, span := otelcommon.Tracer().Start(ctx, "create calculation service", opts...)
	defer span.End()

	// Imitate work
	time.Sleep(30 * time.Millisecond)

	id, err := s.store.CreateCalculation(ctx, postgres.CreateCalculationParams{
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

	if err := s.math.Calculate(ctx, math.Calculation{
		ID:         id,
		Student:    request.Body.Student,
		Expression: request.Body.Expression,
	}); err != nil {
		return api.CreateCalculationdefaultJSONResponse{
			StatusCode: http.StatusInternalServerError,
			Body: api.Error{
				Message: "queue write failure",
			},
		}, nil
	}

	return api.CreateCalculation200JSONResponse{
		Id: id,
	}, nil
}

func (s *service) GetCalculation(ctx context.Context, request api.GetCalculationRequestObject) (api.GetCalculationResponseObject, error) {
	calc, err := s.store.GetCalculation(ctx, request.Uuid)
	if err != nil {
		return api.GetCalculationdefaultJSONResponse{
			StatusCode: http.StatusInternalServerError,
			Body: api.Error{
				Message: "database read failure",
			},
		}, nil
	}

	return api.GetCalculation200JSONResponse{
		Id:         calc.ID,
		Student:    calc.Student,
		Expression: calc.Expression,
		Result:     calc.Result.Float64,
		Created:    calc.Created,
		Completed:  calc.Completed.Time,
	}, nil
}
