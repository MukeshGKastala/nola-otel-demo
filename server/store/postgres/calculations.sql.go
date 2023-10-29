// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.23.0
// source: calculations.sql

package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createCalculation = `-- name: CreateCalculation :one
INSERT INTO calculations (
  student, expression
) VALUES (
  $1, $2
)
RETURNING id, student, expression, result, created, completed
`

type CreateCalculationParams struct {
	Student    string `json:"student"`
	Expression string `json:"expression"`
}

func (q *Queries) CreateCalculation(ctx context.Context, arg CreateCalculationParams) (Calculation, error) {
	row := q.db.QueryRow(ctx, createCalculation, arg.Student, arg.Expression)
	var i Calculation
	err := row.Scan(
		&i.ID,
		&i.Student,
		&i.Expression,
		&i.Result,
		&i.Created,
		&i.Completed,
	)
	return i, err
}

const updateCalculation = `-- name: UpdateCalculation :one
UPDATE calculations
SET
  result = $1,
  completed = $2
WHERE
  id = $3
RETURNING id, student, expression, result, created, completed
`

type UpdateCalculationParams struct {
	Result    pgtype.Text        `json:"result"`
	Completed pgtype.Timestamptz `json:"completed"`
	ID        uuid.UUID          `json:"id"`
}

func (q *Queries) UpdateCalculation(ctx context.Context, arg UpdateCalculationParams) (Calculation, error) {
	row := q.db.QueryRow(ctx, updateCalculation, arg.Result, arg.Completed, arg.ID)
	var i Calculation
	err := row.Scan(
		&i.ID,
		&i.Student,
		&i.Expression,
		&i.Result,
		&i.Created,
		&i.Completed,
	)
	return i, err
}
