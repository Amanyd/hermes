
-- Secrets table for storing sensitive user crendentials

CREATE TABLE IF NOT EXISTS secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, name)
    );

CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id);

DROP INDEX IF EXISTS idx_secrets_user_id;
DROP TABLE IF EXISTS secrets;
