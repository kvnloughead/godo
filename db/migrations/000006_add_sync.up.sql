--- Adds the columns needed for local-first sync. See docs/SPEC.md, "Schema".

--- sync_seq is a per-user counter that goes up by one on every todo write.
--- Writers lock the user's row (SELECT ... FOR UPDATE) to take the next value,
--- so a user's writes commit in seq order.
ALTER TABLE users ADD COLUMN sync_seq bigint NOT NULL DEFAULT 0;

ALTER TABLE todos
  ADD COLUMN uid text,                  --- client ULID; becomes the public ID
  ADD COLUMN updated_at timestamptz NOT NULL DEFAULT NOW(),
  ADD COLUMN deleted_at timestamptz,    --- set on soft delete (a tombstone)
  ADD COLUMN seq bigint NOT NULL DEFAULT 0;

--- Backfill uid for existing rows. New todos get a ULID from the client or the
--- API, but existing rows only need a unique text value, so we use a random
--- UUID (built into Postgres 13+) rather than generating a ULID in SQL. Clients
--- treat uid as an opaque string, so the two formats can coexist.
UPDATE todos SET uid = gen_random_uuid()::text WHERE uid IS NULL;
ALTER TABLE todos ALTER COLUMN uid SET NOT NULL;

--- Existing rows were last modified at some unknown time. created_at is the
--- best available guess, and is earlier than any write made after this
--- migration.
UPDATE todos SET updated_at = created_at;

--- Give existing rows seq values 1..n per user, and set each user's sync_seq
--- to match. Without this, every existing row would have seq = 0, and a full
--- resync (which selects rows with seq > 0) would never return them.
UPDATE todos t
SET seq = numbered.rn
FROM (
  SELECT id, row_number() OVER (PARTITION BY user_id ORDER BY id) AS rn
  FROM todos
) AS numbered
WHERE t.id = numbered.id;

UPDATE users u
SET sync_seq = COALESCE((SELECT max(seq) FROM todos WHERE user_id = u.id), 0);

CREATE UNIQUE INDEX todos_user_uid ON todos (user_id, uid);
CREATE INDEX todos_user_seq ON todos (user_id, seq);
