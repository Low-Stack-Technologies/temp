-- SQLite

-- +migrate Up
ALTER TABLE files
  ALTER COLUMN filename TEXT NOT NULL;

-- +migrate Down

ALTER TABLE files
  ALTER COLUMN filename TEXT;