-- +goose Up
-- 写真メタデータテーブルを追加
CREATE TABLE IF NOT EXISTS photo_metadata (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    course_name TEXT NOT NULL,
    device_photo_id TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    route_stop_id INTEGER,
    photo_synced INTEGER DEFAULT 0,
    taken_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (route_stop_id) REFERENCES route_stops(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_photo_metadata_project_course ON photo_metadata(project_id, course_name);
CREATE INDEX IF NOT EXISTS idx_photo_metadata_stop ON photo_metadata(route_stop_id);
CREATE INDEX IF NOT EXISTS idx_photo_metadata_device_photo ON photo_metadata(project_id, device_photo_id);

-- +goose Down
DROP INDEX IF EXISTS idx_photo_metadata_device_photo;
DROP INDEX IF EXISTS idx_photo_metadata_stop;
DROP INDEX IF EXISTS idx_photo_metadata_project_course;
DROP TABLE IF EXISTS photo_metadata;
