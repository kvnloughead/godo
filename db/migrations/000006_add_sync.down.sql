--- Soft-deleted rows must be removed before dropping deleted_at. Otherwise
--- every deleted todo would reappear as a live one.
DELETE FROM todos WHERE deleted_at IS NOT NULL;

DROP INDEX IF EXISTS todos_user_seq;
DROP INDEX IF EXISTS todos_user_uid;

ALTER TABLE todos
  DROP COLUMN IF EXISTS seq,
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS uid;

ALTER TABLE users DROP COLUMN IF EXISTS sync_seq;
