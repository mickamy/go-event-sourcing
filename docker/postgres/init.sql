CREATE DATABASE ges;

\c ges

CREATE TABLE IF NOT EXISTS events
(
    stream_id  TEXT        NOT NULL,
    version    BIGINT      NOT NULL,
    event_id   UUID                 DEFAULT gen_random_uuid(),
    event_type TEXT        NOT NULL,
    payload    JSONB       NOT NULL,
    metadata   JSONB       NOT NULL DEFAULT '{}'::jsonb,
    at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (stream_id, version),
    UNIQUE (event_id)
);

CREATE TABLE IF NOT EXISTS snapshots
(
    stream_id TEXT PRIMARY KEY,
    version   BIGINT      NOT NULL,
    state     JSONB       NOT NULL,
    at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
