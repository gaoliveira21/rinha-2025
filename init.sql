CREATE UNLOGGED TABLE payments_default (
  correlationId UUID PRIMARY KEY,
  amount DECIMAL NOT NULL,
  requested_at TIMESTAMP NOT NULL
);

CREATE INDEX payments_default_requested_at ON payments_default (requested_at);

CREATE UNLOGGED TABLE payments_fallback (
  correlationId UUID PRIMARY KEY,
  amount DECIMAL NOT NULL,
  requested_at TIMESTAMP NOT NULL
);

CREATE INDEX payments_fallback_requested_at ON payments_fallback (requested_at);
