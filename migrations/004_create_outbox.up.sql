
CREATE TABLE outbox_events (
    id UUID PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    attempts INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,

    CONSTRAINT outbox_attempts_valid
        CHECK (attempts >= 0),

    CONSTRAINT outbox_event_type_valid
        CHECK (event_type <> '')
);

CREATE INDEX outbox_pending_idx
ON outbox_events (next_attempt_at, occurred_at)
WHERE published_at IS NULL;
