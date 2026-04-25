CREATE TABLE IF NOT EXISTS deployments (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    source_type  TEXT NOT NULL CHECK(source_type IN ('git', 'upload')),
    source_url   TEXT,
    status       TEXT NOT NULL DEFAULT 'pending',
    image_tag    TEXT,
    container_id TEXT,
    host_port    INTEGER,
    live_url     TEXT,
    error        TEXT,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS logs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    stream        TEXT NOT NULL CHECK(stream IN ('build', 'deploy', 'system')),
    line          TEXT NOT NULL,
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_logs_deployment_id ON logs(deployment_id);