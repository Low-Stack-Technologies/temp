-- SQLite

-- +migrate Up
ALTER TABLE upload_groups ADD COLUMN archive_name TEXT;

-- +migrate Down
-- SQLite does not support dropping a column without rebuilding the table.
