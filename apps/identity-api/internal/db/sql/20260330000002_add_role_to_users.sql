-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT DEFAULT 'user';

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS role;
