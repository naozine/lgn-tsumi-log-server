-- +goose Up
-- デバイス管理テーブルを追加
CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    device_id TEXT NOT NULL,
    device_name TEXT,
    course_name TEXT,
    last_seen_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    UNIQUE(project_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_devices_project ON devices(project_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_id ON devices(project_id, device_id);

-- +goose Down
DROP INDEX IF EXISTS idx_devices_device_id;
DROP INDEX IF EXISTS idx_devices_project;
DROP TABLE IF EXISTS devices;
