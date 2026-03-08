DROP INDEX IF EXISTS idx_relays_next_run_at;
DROP INDEX IF EXISTS idx_relays_trigger_type;

ALTER TABLE relays
  DROP COLUMN IF EXISTS last_run_at,
  DROP COLUMN IF EXISTS next_run_at,
  DROP COLUMN IF EXISTS trigger_config,
  DROP COLUMN IF EXISTS trigger_type;
