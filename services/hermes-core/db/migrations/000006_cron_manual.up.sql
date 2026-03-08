ALTER TABLE relays
  ADD COLUMN IF NOT EXISTS trigger_type TEXT NOT NULL DEFAULT 'webhook',
  ADD COLUMN IF NOT EXISTS trigger_config JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE relays
  ADD COLUMN IF NOT EXISTS next_run_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS last_run_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_relays_trigger_type ON relays(trigger_type);
CREATE INDEX IF NOT EXISTS idx_relays_next_run_at ON relays(next_run_at);
