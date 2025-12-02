-- +goose Up
ALTER TABLE projects ADD COLUMN judge_stay_time_minutes INTEGER NOT NULL DEFAULT 0;
ALTER TABLE projects ADD COLUMN judge_speed_limit_kmh REAL NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE projects DROP COLUMN judge_stay_time_minutes;
ALTER TABLE projects DROP COLUMN judge_speed_limit_kmh;