-- Files table to store metadata about uploaded files
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    filename TEXT NOT NULL,
    size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    storage_path TEXT NOT NULL,
    upload_method TEXT NOT NULL CHECK(upload_method IN ('http', 'tcp')),
    uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ttl_seconds INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    checksum TEXT NOT NULL
);

-- Index on expires_at for efficient cleanup queries
CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at);

-- Index on upload_method for analytics
CREATE INDEX IF NOT EXISTS idx_files_upload_method ON files(upload_method);
