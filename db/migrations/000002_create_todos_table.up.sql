CREATE TABLE IF NOT EXISTS todos (
  id bigserial PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
  text text NOT NULL,
  contexts text[] NOT NULL,
  projects text[] NOT NULL,
  priority text CHECK (priority IN ('', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z')),
  completed BOOLEAN NOT NULL,
  version integer NOT NULL DEFAULT 1
);