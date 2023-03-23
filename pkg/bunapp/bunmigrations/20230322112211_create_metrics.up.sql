DO $$ BEGIN
  CREATE TYPE public.metric_instrument_enum AS ENUM (
    'gauge',
    'additive',
    'counter',
    'histogram'
  );
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

--bun:split


CREATE TABLE metrics (
  id int8 PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
  project_id int4,

  name varchar(1000) NOT NULL,
  description varchar(1000),
  unit varchar(100),
  instrument metric_instrument_enum NOT NULL,

  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz
);

--bun:split

CREATE UNIQUE INDEX metrics_project_id_name_unq
ON metrics (project_id, name);