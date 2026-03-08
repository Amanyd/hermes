CREATE TABLE IF NOT EXISTS connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    account_email TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expiry TIMESTAMP NOT NULL,
    scopes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, provider, account_email)
);

CREATE INDEX IF NOT EXISTS idx_connections_user_id ON connections(user_id);
CREATE INDEX IF NOT EXISTS idx_connections_provider ON connections(provider);
