-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (name, api_key, arrival_threshold_meters, judge_stay_time_minutes, judge_speed_limit_kmh)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = ? LIMIT 1;

-- name: UpdateProject :one
UPDATE projects
SET name = ?, arrival_threshold_meters = ?, judge_stay_time_minutes = ?, judge_speed_limit_kmh = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;

-- name: UpdateProjectCSV :one
UPDATE projects
SET csv_filename = ?, csv_imported_at = ?, csv_row_count = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: CreateRouteStop :exec
INSERT INTO route_stops (
    project_id, course_name, sequence, arrival_time, stop_name,
    address, latitude, longitude, stay_minutes, weight_kg,
    phone_number, note1, note2, note3,
    desired_time_start, desired_time_end
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: DeleteRouteStopsByProject :exec
DELETE FROM route_stops WHERE project_id = ?;

-- name: ListCoursesByProject :many
SELECT DISTINCT course_name FROM route_stops WHERE project_id = ? ORDER BY course_name;

-- name: ListRouteStopsByCourse :many
SELECT * FROM route_stops WHERE project_id = ? AND course_name = ? ORDER BY arrival_time;

-- name: GetRouteStopByID :one
SELECT * FROM route_stops WHERE id = ? LIMIT 1;

-- name: CreateLocationLog :exec
INSERT INTO location_logs (
    project_id, course_name, latitude, longitude, timestamp,
    accuracy, speed, bearing, battery_level
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListLocationLogsByCourse :many
SELECT * FROM location_logs
WHERE project_id = ? AND course_name = ?
ORDER BY timestamp;

-- name: ListLocationLogsByCourseDesc :many
SELECT * FROM location_logs
WHERE project_id = ? AND course_name = ?
ORDER BY timestamp DESC;

-- name: GetLatestLocationByCourse :one
SELECT * FROM location_logs
WHERE project_id = ? AND course_name = ?
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetProjectByAPIKey :one
SELECT * FROM projects WHERE api_key = ? LIMIT 1;

-- name: UpdateProjectAPIKey :one
UPDATE projects
SET api_key = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteLocationLogsByCourse :exec
DELETE FROM location_logs
WHERE project_id = ? AND course_name = ?;

-- name: CreatePhotoMetadata :one
INSERT INTO photo_metadata (
    project_id, course_name, device_photo_id, latitude, longitude, route_stop_id, taken_at
)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPhotoMetadataByDeviceID :one
SELECT * FROM photo_metadata
WHERE project_id = ? AND device_photo_id = ?
LIMIT 1;

-- name: ListPhotoMetadataByCourse :many
SELECT * FROM photo_metadata
WHERE project_id = ? AND course_name = ?
ORDER BY taken_at;

-- name: ListPhotoMetadataByStop :many
SELECT * FROM photo_metadata
WHERE route_stop_id = ?
ORDER BY taken_at;

-- name: UpdatePhotoSynced :exec
UPDATE photo_metadata
SET photo_synced = 1
WHERE id = ?;

-- name: CreateDevice :one
INSERT INTO devices (project_id, device_id, device_name)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetDeviceByDeviceID :one
SELECT * FROM devices
WHERE project_id = ? AND device_id = ?
LIMIT 1;

-- name: ListDevicesByProject :many
SELECT * FROM devices
WHERE project_id = ?
ORDER BY created_at DESC;

-- name: UpdateDeviceCourseName :one
UPDATE devices
SET course_name = ?, last_seen_at = CURRENT_TIMESTAMP
WHERE project_id = ? AND device_id = ?
RETURNING *;

-- name: UpdateDeviceLastSeen :exec
UPDATE devices
SET last_seen_at = CURRENT_TIMESTAMP
WHERE project_id = ? AND device_id = ?;

-- name: DeleteDevice :exec
DELETE FROM devices
WHERE project_id = ? AND device_id = ?;
