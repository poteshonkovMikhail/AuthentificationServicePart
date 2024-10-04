CREATE TABLE IF NOT EXISTS refresh_tokens (
    user_guid VARCHAR UNIQUE,
    token_hash TEXT,
    client_ip VARCHAR,
    expires_at TIMESTAMP
);