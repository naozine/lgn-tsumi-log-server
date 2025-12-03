package testdata

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/naozine/project_crud_with_auth_tmpl/internal/database"
)

// Generator はテストデータ生成器
type Generator struct {
	DB         *database.Queries
	Directions *DirectionsClient
	Interval   int // ログ生成間隔（秒）
}

// NewGenerator は新しいGeneratorを作成する
func NewGenerator(db *database.Queries, apiKey string, interval int) *Generator {
	return &Generator{
		DB:         db,
		Directions: NewDirectionsClient(apiKey),
		Interval:   interval,
	}
}

// LocationLog は生成される位置情報ログ
type LocationLog struct {
	Latitude     float64
	Longitude    float64
	Timestamp    time.Time
	Accuracy     float64
	Speed        float64
	Bearing      float64
	BatteryLevel int
}

// GenerateResult は生成結果
type GenerateResult struct {
	DeletedLogs   int
	GeneratedLogs int
	RouteSegments int
}

// Generate はコースのテストデータを生成する
func (g *Generator) Generate(ctx context.Context, projectID int64, courseName string, baseDate time.Time) (*GenerateResult, error) {
	result := &GenerateResult{}

	// 1. 停車地一覧を取得
	stops, err := g.DB.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
		ProjectID:  projectID,
		CourseName: courseName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get route stops: %w", err)
	}

	if len(stops) < 2 {
		return nil, fmt.Errorf("at least 2 stops required, got %d", len(stops))
	}

	// 2. 既存ログを削除
	if err := g.DB.DeleteLocationLogsByCourse(ctx, database.DeleteLocationLogsByCourseParams{
		ProjectID:  projectID,
		CourseName: courseName,
	}); err != nil {
		return nil, fmt.Errorf("failed to delete existing logs: %w", err)
	}

	// 3. 各停車地ペアに対してルートを取得しログを生成
	var allLogs []LocationLog
	batteryLevel := 100

	for i := 0; i < len(stops)-1; i++ {
		stopA := stops[i]
		stopB := stops[i+1]

		// 座標がない場合はスキップ
		if !stopA.Latitude.Valid || !stopA.Longitude.Valid ||
			!stopB.Latitude.Valid || !stopB.Longitude.Valid {
			fmt.Printf("  [%d/%d] %s → %s ... スキップ（座標なし）\n", i+1, len(stops)-1, stopA.StopName, stopB.StopName)
			continue
		}

		origin := LatLng{Lat: stopA.Latitude.Float64, Lng: stopA.Longitude.Float64}
		destination := LatLng{Lat: stopB.Latitude.Float64, Lng: stopB.Longitude.Float64}

		// ルートを取得
		route, err := g.Directions.GetRoute(origin, destination)
		if err != nil {
			fmt.Printf("  [%d/%d] %s → %s ... エラー: %v\n", i+1, len(stops)-1, stopA.StopName, stopB.StopName, err)
			continue
		}

		fmt.Printf("  [%d/%d] %s → %s ... OK (%dポイント)\n", i+1, len(stops)-1, stopA.StopName, stopB.StopName, len(route.Points))
		result.RouteSegments++

		// 出発時刻と到着時刻を計算
		departureTime := parseArrivalTime(stopA.ArrivalTime.String, baseDate)
		if stopA.StayMinutes.Valid {
			departureTime = departureTime.Add(time.Duration(stopA.StayMinutes.Int64) * time.Minute)
		}
		arrivalTime := parseArrivalTime(stopB.ArrivalTime.String, baseDate)

		// 滞在ログを生成（最初の停車地のみ、最初のループで）
		if i == 0 {
			stayLogs := g.generateStayLogs(stopA, baseDate, &batteryLevel)
			allLogs = append(allLogs, stayLogs...)
		}

		// 移動ログを生成
		moveLogs := g.generateMoveLogs(route.Points, departureTime, arrivalTime, &batteryLevel)
		allLogs = append(allLogs, moveLogs...)

		// 到着地の滞在ログを生成
		stayLogs := g.generateStayLogs(stopB, baseDate, &batteryLevel)
		allLogs = append(allLogs, stayLogs...)
	}

	// 4. DBに挿入
	for _, log := range allLogs {
		err := g.DB.CreateLocationLog(ctx, database.CreateLocationLogParams{
			ProjectID:    projectID,
			CourseName:   courseName,
			Latitude:     log.Latitude,
			Longitude:    log.Longitude,
			Timestamp:    log.Timestamp,
			Accuracy:     sql.NullFloat64{Float64: log.Accuracy, Valid: true},
			Speed:        sql.NullFloat64{Float64: log.Speed, Valid: true},
			Bearing:      sql.NullFloat64{Float64: log.Bearing, Valid: log.Bearing >= 0},
			BatteryLevel: sql.NullInt64{Int64: int64(log.BatteryLevel), Valid: true},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert log: %w", err)
		}
	}

	result.GeneratedLogs = len(allLogs)
	return result, nil
}

// generateStayLogs は停車地での滞在ログを生成する
func (g *Generator) generateStayLogs(stop database.RouteStop, baseDate time.Time, batteryLevel *int) []LocationLog {
	var logs []LocationLog

	if !stop.Latitude.Valid || !stop.Longitude.Valid {
		return logs
	}

	arrivalTime := parseArrivalTime(stop.ArrivalTime.String, baseDate)
	stayMinutes := int64(0)
	if stop.StayMinutes.Valid {
		stayMinutes = stop.StayMinutes.Int64
	}

	// 滞在中は interval 秒ごとにログを生成
	numLogs := int(stayMinutes*60) / g.Interval
	if numLogs < 1 {
		numLogs = 1
	}

	for i := 0; i <= numLogs; i++ {
		timestamp := arrivalTime.Add(time.Duration(i*g.Interval) * time.Second)

		// GPSノイズを付加（±5m程度）
		lat := stop.Latitude.Float64 + (rand.Float64()-0.5)*0.0001
		lng := stop.Longitude.Float64 + (rand.Float64()-0.5)*0.0001

		logs = append(logs, LocationLog{
			Latitude:     lat,
			Longitude:    lng,
			Timestamp:    timestamp,
			Accuracy:     5 + rand.Float64()*10, // 5-15m
			Speed:        rand.Float64() * 3,    // 0-3 km/h
			Bearing:      -1,                    // 無効
			BatteryLevel: *batteryLevel,
		})

		// バッテリーを少し減らす
		if rand.Float64() < 0.1 && *batteryLevel > 0 {
			*batteryLevel--
		}
	}

	return logs
}

// generateMoveLogs は移動中のログを生成する
func (g *Generator) generateMoveLogs(points []LatLng, departureTime, arrivalTime time.Time, batteryLevel *int) []LocationLog {
	var logs []LocationLog

	if len(points) < 2 {
		return logs
	}

	totalDuration := arrivalTime.Sub(departureTime)
	if totalDuration <= 0 {
		return logs
	}

	// 総距離を計算
	totalDistance := 0.0
	for i := 1; i < len(points); i++ {
		totalDistance += haversineDistance(points[i-1], points[i])
	}

	if totalDistance == 0 {
		return logs
	}

	// 各ポイントに時刻を割り当て
	currentDistance := 0.0
	for i := 0; i < len(points); i++ {
		var timestamp time.Time
		var speed float64
		var bearing float64

		if i == 0 {
			timestamp = departureTime
		} else {
			segmentDistance := haversineDistance(points[i-1], points[i])
			currentDistance += segmentDistance
			ratio := currentDistance / totalDistance
			timestamp = departureTime.Add(time.Duration(float64(totalDuration) * ratio))

			// 速度を計算（km/h）
			if i > 0 {
				timeDiff := timestamp.Sub(logs[len(logs)-1].Timestamp).Hours()
				if timeDiff > 0 {
					speed = segmentDistance / timeDiff
					// ランダム変動を加える（±5%）
					speed *= 0.95 + rand.Float64()*0.1
				}
			}

			// 方位角を計算
			bearing = calculateBearing(points[i-1], points[i])
		}

		// GPSノイズを付加
		lat := points[i].Lat + (rand.Float64()-0.5)*0.00005
		lng := points[i].Lng + (rand.Float64()-0.5)*0.00005

		logs = append(logs, LocationLog{
			Latitude:     lat,
			Longitude:    lng,
			Timestamp:    timestamp,
			Accuracy:     5 + rand.Float64()*10,
			Speed:        speed,
			Bearing:      bearing,
			BatteryLevel: *batteryLevel,
		})

		// バッテリーを少し減らす
		if rand.Float64() < 0.05 && *batteryLevel > 0 {
			*batteryLevel--
		}
	}

	return logs
}

// parseArrivalTime は "HH:MM" 形式の時刻をパースする
func parseArrivalTime(timeStr string, baseDate time.Time) time.Time {
	if timeStr == "" {
		return baseDate
	}

	var hour, minute int
	fmt.Sscanf(timeStr, "%d:%d", &hour, &minute)

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), hour, minute, 0, 0, jst)
}

// haversineDistance は2点間の距離をキロメートルで計算する
func haversineDistance(p1, p2 LatLng) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := p1.Lat * math.Pi / 180
	lat2Rad := p2.Lat * math.Pi / 180
	dLat := (p2.Lat - p1.Lat) * math.Pi / 180
	dLng := (p2.Lng - p1.Lng) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// calculateBearing は2点間の方位角を計算する（0-360度）
func calculateBearing(p1, p2 LatLng) float64 {
	lat1Rad := p1.Lat * math.Pi / 180
	lat2Rad := p2.Lat * math.Pi / 180
	dLng := (p2.Lng - p1.Lng) * math.Pi / 180

	y := math.Sin(dLng) * math.Cos(lat2Rad)
	x := math.Cos(lat1Rad)*math.Sin(lat2Rad) - math.Sin(lat1Rad)*math.Cos(lat2Rad)*math.Cos(dLng)

	bearing := math.Atan2(y, x) * 180 / math.Pi
	return math.Mod(bearing+360, 360)
}
