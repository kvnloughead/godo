CREATE TYPE priority_level AS ENUM ('A', 'B', 'C');

CREATE TABLE IF NOT EXISTS todos (
  id bigserial PRIMARY KEY,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  title text NOT NULL,
  contexts text[] NOT NULL,
  projects text[] NOT NULL,
  priority priority_level NOT NULL,
  completed BOOLEAN NOT NULL,
  metadata jsonb,
  version integer NOT NULL DEFAULT 1
);