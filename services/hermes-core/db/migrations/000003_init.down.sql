DROP INDEX IF EXISTS idx_execution_logs_events_id;
ALTER TABLE execution_logs DROP COLUMN IF EXISTS event_id;
