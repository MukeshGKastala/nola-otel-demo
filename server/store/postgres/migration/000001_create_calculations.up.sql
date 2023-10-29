CREATE TABLE calculations (
  id SERIAL PRIMARY KEY, 
  student VARCHAR NOT NULL,
  expression VARCHAR NOT NULL,
  result VARCHAR,
  created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed TIMESTAMPTZ
);
