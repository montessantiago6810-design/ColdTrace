-- +goose Up
-- TimescaleDB is enabled where the server provides it (the Compose image
-- does). Without it the schema still applies; hypertables arrive with the
-- readings migration, which falls back to native partitioning (ADR 0001).
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'timescaledb') THEN
        CREATE EXTENSION IF NOT EXISTS timescaledb;
    END IF;
END
$$;
-- +goose StatementEnd

CREATE TABLE organizations (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       text        NOT NULL,
    name       text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT organizations_slug_key UNIQUE (slug),
    CONSTRAINT organizations_slug_format CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,38}[a-z0-9])?$'),
    CONSTRAINT organizations_name_length CHECK (char_length(btrim(name)) BETWEEN 1 AND 120)
);

COMMENT ON TABLE organizations IS 'Tenants. Every org-owned row references organizations.id.';

-- +goose Down
DROP TABLE organizations;
