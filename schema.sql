-- Drop table if it exists (for idempotency)
DROP TABLE IF EXISTS delegations;

-- Main delegations table
CREATE TABLE delegations (
    id BIGINT PRIMARY KEY,
    delegator VARCHAR(36) NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    amount BIGINT NOT NULL,
    level INTEGER NOT NULL
);

CREATE INDEX idx_delegations_timestamp ON delegations(timestamp);
