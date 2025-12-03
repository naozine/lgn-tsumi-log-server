-- +goose Up
-- location_logsにdevice_idカラムを追加（将来のデバイス単位分析用）
ALTER TABLE location_logs ADD COLUMN device_id TEXT;

-- device_idで検索する場合のインデックス
CREATE INDEX IF NOT EXISTS idx_location_logs_device_id ON location_logs(device_id);

-- +goose Down
DROP INDEX IF EXISTS idx_location_logs_device_id;
-- SQLiteはALTER TABLE DROP COLUMNをサポートしないため、カラム削除は省略
