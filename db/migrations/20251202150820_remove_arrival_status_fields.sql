-- +goose Up
-- route_stops から status, actual_arrival_time, actual_departure_time を削除
-- SQLite では ALTER TABLE DROP COLUMN がサポートされていないため、テーブルを再作成

-- 一時テーブルを作成
CREATE TABLE route_stops_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    course_name TEXT NOT NULL,
    sequence TEXT NOT NULL,
    arrival_time TEXT,
    stop_name TEXT NOT NULL,
    address TEXT,
    latitude REAL,
    longitude REAL,
    stay_minutes INTEGER DEFAULT 0,
    weight_kg INTEGER DEFAULT 0,
    phone_number TEXT,
    note1 TEXT,
    note2 TEXT,
    note3 TEXT,
    desired_time_start TEXT,
    desired_time_end TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- データを移行
INSERT INTO route_stops_new (
    id, project_id, course_name, sequence, arrival_time, stop_name,
    address, latitude, longitude, stay_minutes, weight_kg,
    phone_number, note1, note2, note3, desired_time_start, desired_time_end,
    created_at, updated_at
)
SELECT
    id, project_id, course_name, sequence, arrival_time, stop_name,
    address, latitude, longitude, stay_minutes, weight_kg,
    phone_number, note1, note2, note3, desired_time_start, desired_time_end,
    created_at, updated_at
FROM route_stops;

-- 旧テーブルを削除
DROP TABLE route_stops;

-- 新テーブルをリネーム
ALTER TABLE route_stops_new RENAME TO route_stops;

-- インデックスを再作成
CREATE INDEX idx_route_stops_project ON route_stops(project_id);
CREATE INDEX idx_route_stops_course ON route_stops(course_name);

-- +goose Down
-- status, actual_arrival_time, actual_departure_time を復元
ALTER TABLE route_stops ADD COLUMN status TEXT DEFAULT '未訪問';
ALTER TABLE route_stops ADD COLUMN actual_arrival_time TEXT;
ALTER TABLE route_stops ADD COLUMN actual_departure_time TEXT;
