-- name: UpsertProjectLogisticsSettings :one
INSERT INTO project_logistics_settings (
    project_id, csv_filename, csv_imported_at, csv_row_count, arrival_threshold_meters, updated_at
) VALUES (
    ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
ON CONFLICT(project_id) DO UPDATE SET
    csv_filename = excluded.csv_filename,
    csv_imported_at = excluded.csv_imported_at,
    csv_row_count = excluded.csv_row_count,
    arrival_threshold_meters = excluded.arrival_threshold_meters,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetProjectLogisticsSettings :one
SELECT * FROM project_logistics_settings WHERE project_id = ? LIMIT 1;

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
