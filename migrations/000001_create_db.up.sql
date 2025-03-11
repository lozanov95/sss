CREATE EXTENSION pgcrypto;

CREATE TABLE IF NOT EXISTS secrets (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    expires_at timestamp(0) with time zone NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    passphrase bytea
);

CREATE INDEX IF NOT EXISTS key_idx ON secrets USING Hash (key);