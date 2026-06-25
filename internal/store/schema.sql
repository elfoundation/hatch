CREATE TABLE IF NOT EXISTS endpoints (
    id         TEXT NOT NULL PRIMARY KEY,
    url        TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    mock_status INTEGER,
    mock_headers TEXT,
    mock_body BLOB
);
CREATE TABLE IF NOT EXISTS requests (
    id          TEXT NOT NULL PRIMARY KEY,
    endpoint_id TEXT NOT NULL REFERENCES endpoints(id),
    method      TEXT NOT NULL,
    path        TEXT NOT NULL,
    headers     TEXT NOT NULL DEFAULT '{}',
    query       TEXT NOT NULL DEFAULT '',
    body        BLOB,
    created_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_requests_endpoint_id ON requests(endpoint_id);
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at);
