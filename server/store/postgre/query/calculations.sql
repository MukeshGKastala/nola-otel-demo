-- name: CreateCalculation :one
INSERT INTO calculations (
  student,
  expression,
) VALUES (
  $1, $2
) RETURNING id;