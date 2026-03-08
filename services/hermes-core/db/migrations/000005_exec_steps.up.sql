CREATE TABLE IF NOT EXISTS executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    relay_id UUID NOT NULL REFERENCES relays(id) ON DELETE CASCADE,
    event_id TEXT,
    status TEXT NOT NULL,
    trigger_payload JSONB,
    error_message TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS execution_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE,
    order_index INT NOT NULL,
    action_type TEXT NOT NULL,
    status TEXT NOT NULL,
    input JSONB,
    output JSONB,
    error_message TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_executions_relay_id ON executions(relay_id);
CREATE INDEX IF NOT EXISTS idx_executions_started_at ON executions(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_executions_event_id ON executions(event_id);

CREATE INDEX IF NOT EXISTS idx_execution_steps_execution_id ON execution_steps(execution_id);
CREATE INDEX IF NOT EXISTS idx_execution_steps_order_index ON execution_steps(execution_id, order_index);
