CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE calculations (
  id uuid DEFAULT uuid_generate_v4 (), 
  student VARCHAR NOT NULL,
  expression VARCHAR NOT NULL,
  result DOUBLE PRECISION,
  created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  completed TIMESTAMPTZ,
  PRIMARY KEY (id)
);
