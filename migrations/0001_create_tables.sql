CREATE TABLE events (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    source      text        NOT NULL,
    headers     jsonb       NOT NULL,
    body        bytea       NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE deliveries (
    id              bigint      GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id        uuid        NOT NULL REFERENCES events(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'succeeded', 'dead')),
    attempts         int        NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT NOW(),
    last_error      text        NULL,
    created_at      timestamptz NOT NULL DEFAULT NOW(),
    updated_at      timestamptz NOT NULL DEFAULT NOW()
);