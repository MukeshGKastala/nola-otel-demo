-- name: CreateCalculation :one
INSERT INTO calculations (
  student, expression
) VALUES (
  $1, $2
)
RETURNING id;

-- name: GetCalculation :one
SELECT * FROM calculations
WHERE id = $1;

-- name: UpdateCalculation :one
UPDATE calculations
SET
  result = $1,
  completed = $2
WHERE
  id = $3
RETURNING *;