-- +goose Up
ALTER TABLE projects ADD COLUMN api_key TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_api_key ON projects (api_key);

-- +goose Down
DROP INDEX IF EXISTS idx_projects_api_key;
ALTER TABLE projects DROP COLUMN api_key;