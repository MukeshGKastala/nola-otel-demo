-- name: CreateCalculation :one
INSERT INTO calculations (
  student, expression
) VALUES (
  $1, $2
)
RETURNING *;

-- name: UpdateCalculation :one
UPDATE calculations
SET
  result = $1,
  completed = $2
WHERE
  id = $3
RETURNING *;