-- SQLite

-- +migrate Up
CREATE TABLE files_new (
  id TEXT PRIMARY KEY,
  filename TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL
);
INSERT INTO files_new SELECT * FROM files;
DROP TABLE files;
ALTER TABLE files_new RENAME TO files;

-- +migrate Down
CREATE TABLE files_new (
  id TEXT PRIMARY KEY,
  filename TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  expires_at DATETIME NOT NULL
);
INSERT INTO files_new SELECT * FROM files;
DROP TABLE files;
ALTER TABLE files_new RENAME TO files;