CREATE TABLE IF NOT EXISTS client (
    client_id VARCHAR(255) PRIMARY KEY,
    capacity INTEGER NOT NULL,
    tokens INTEGER NOT NULL
);