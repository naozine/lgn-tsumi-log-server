-- 物流案件（プロジェクト）管理テーブル
-- projectsテーブルを物流用に拡張した定義で上書き
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    api_key TEXT UNIQUE NOT NULL DEFAULT '',
    csv_filename TEXT,
    csv_imported_at DATETIME,
    csv_row_count INTEGER DEFAULT 0,
    arrival_threshold_meters INTEGER DEFAULT 100,
    judge_stay_time_minutes INTEGER DEFAULT 0,
    judge_speed_limit_kmh REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

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

-- 写真メタデータ（実データは後で同期）
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

-- デバイス管理
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
