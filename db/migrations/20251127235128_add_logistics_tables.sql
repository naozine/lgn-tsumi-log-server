-- +goose Up
-- projectsテーブルに物流関連カラムを追加
ALTER TABLE projects ADD COLUMN csv_filename TEXT;
ALTER TABLE projects ADD COLUMN csv_imported_at DATETIME;
ALTER TABLE projects ADD COLUMN csv_row_count INTEGER DEFAULT 0;
ALTER TABLE projects ADD COLUMN arrival_threshold_meters INTEGER DEFAULT 100;
ALTER TABLE projects ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;

-- 配送停車地データ
CREATE TABLE IF NOT EXISTS route_stops (
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
    status TEXT DEFAULT '未訪問',
    actual_arrival_time TEXT,
    actual_departure_time TEXT,
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

CREATE INDEX IF NOT EXISTS idx_route_stops_project ON route_stops(project_id);
CREATE INDEX IF NOT EXISTS idx_route_stops_course ON route_stops(course_name);

-- 位置情報ログ
CREATE TABLE IF NOT EXISTS location_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    course_name TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    timestamp DATETIME NOT NULL,
    accuracy REAL,
    speed REAL,
    bearing REAL,
    battery_level INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_location_logs_project_course ON location_logs(project_id, course_name);
CREATE INDEX IF NOT EXISTS idx_location_logs_timestamp ON location_logs(timestamp);

-- +goose Down
-- projectsテーブルのカラム削除はSQLiteではALTER TABLE DROP COLUMNが限定的、
-- またはテーブル再作成が必要になるため、ここでは行わない。
-- 通常はDOWNでカラムを削除すべきだが、開発用途でDB初期化が前提ならスキップ可。
-- 派生プロジェクト側でprojectsテーブルを変更しているため、ベースのDOWNと競合する可能性もある。
DROP TABLE IF EXISTS location_logs;
DROP TABLE IF EXISTS route_stops;
