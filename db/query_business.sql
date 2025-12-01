-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: CreateProject :one
INSERT INTO projects (name, api_key, arrival_threshold_meters)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = ? LIMIT 1;

-- name: UpdateProject :one
UPDATE projects
SET name = ?, arrival_threshold_meters = ?, updated_at = CURRENT_TIMESTAMP
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
    status, phone_number, note1, note2, note3,
    desired_time_start, desired_time_end
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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

-- name: GetLatestLocationByCourse :one
SELECT * FROM location_logs
WHERE project_id = ? AND course_name = ?
ORDER BY timestamp DESC
LIMIT 1;

-- name: UpdateRouteStopStatus :exec
UPDATE route_stops
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: ResetRouteStopsStatusByCourse :exec
UPDATE route_stops 
SET status = ?, actual_arrival_time = NULL, actual_departure_time = NULL, updated_at = CURRENT_TIMESTAMP
WHERE project_id = ? AND course_name = ?;

-- name: UpdateRouteStopArrival :exec
UPDATE route_stops 
SET status = ?, actual_arrival_time = ?, actual_departure_time = NULL, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: UpdateRouteStopDeparture :exec
UPDATE route_stops 
SET actual_departure_time = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: ClearRouteStopDeparture :exec
UPDATE route_stops 
SET actual_departure_time = NULL, updated_at = CURRENT_TIMESTAMP 
WHERE id = ?;

-- name: GetProjectByAPIKey :one
SELECT * FROM projects WHERE api_key = ? LIMIT 1;

-- name: UpdateProjectAPIKey :one
UPDATE projects
SET api_key = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

