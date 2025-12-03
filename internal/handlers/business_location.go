package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/database"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/geo"
)

type LocationHandler struct {
	DB *database.Queries
}

func NewLocationHandler(db *database.Queries) *LocationHandler {
	return &LocationHandler{DB: db}
}

// リクエスト用の構造体
type LocationData struct {
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	Timestamp    string   `json:"timestamp"`
	Accuracy     *float64 `json:"accuracy,omitempty"`
	Speed        *float64 `json:"speed,omitempty"`
	Bearing      *float64 `json:"bearing,omitempty"`
	BatteryLevel *int64   `json:"battery_level,omitempty"`
}

type LocationRequest struct {
	DeviceID  string         `json:"device_id"`
	Locations []LocationData `json:"locations"`
}

// デバイス登録API用の構造体
type DeviceRegisterRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type DeviceRegisterResponse struct {
	Success    bool    `json:"success"`
	DeviceID   string  `json:"device_id,omitempty"`
	CourseName *string `json:"course_name"`
	Message    string  `json:"message,omitempty"`
	Error      string  `json:"error,omitempty"`
}

// レスポンス用の構造体
type LocationResponse struct {
	Success  bool   `json:"success"`
	Recorded int    `json:"recorded,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// POST /api/v1/devices
func (h *LocationHandler) RegisterDevice(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. X-Project-Api-Key ヘッダーからAPIキーを取得
	apiKey := c.Request().Header.Get("X-Project-Api-Key")
	if apiKey == "" {
		return c.JSON(http.StatusUnauthorized, DeviceRegisterResponse{
			Success: false,
			Error:   "API key is required",
		})
	}

	// 2. APIキーでプロジェクトを検索し認証
	project, err := h.DB.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, DeviceRegisterResponse{
				Success: false,
				Error:   "Invalid API key",
			})
		}
		log.Printf("Database error during API key validation: %v", err)
		return c.JSON(http.StatusInternalServerError, DeviceRegisterResponse{
			Success: false,
			Error:   "Internal server error during API key validation",
		})
	}

	var req DeviceRegisterRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, DeviceRegisterResponse{
			Success: false,
			Error:   "Invalid request format",
		})
	}

	// バリデーション
	if req.DeviceID == "" {
		return c.JSON(http.StatusBadRequest, DeviceRegisterResponse{
			Success: false,
			Error:   "device_id is required",
		})
	}

	// 既存デバイスを確認
	existingDevice, err := h.DB.GetDeviceByDeviceID(ctx, database.GetDeviceByDeviceIDParams{
		ProjectID: project.ID,
		DeviceID:  req.DeviceID,
	})
	if err == nil {
		// 既に登録済み - 現在の情報を返す
		var courseName *string
		if existingDevice.CourseName.Valid {
			courseName = &existingDevice.CourseName.String
		}
		msg := "Device already registered."
		if courseName != nil {
			msg = fmt.Sprintf("Device already registered. Assigned to course: %s", *courseName)
		} else {
			msg = "Device already registered. No course assigned yet."
		}
		return c.JSON(http.StatusOK, DeviceRegisterResponse{
			Success:    true,
			DeviceID:   existingDevice.DeviceID,
			CourseName: courseName,
			Message:    msg,
		})
	}

	if err != sql.ErrNoRows {
		log.Printf("Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, DeviceRegisterResponse{
			Success: false,
			Error:   "Failed to check existing device",
		})
	}

	// 新規デバイスを登録
	deviceName := sql.NullString{String: req.DeviceName, Valid: req.DeviceName != ""}
	device, err := h.DB.CreateDevice(ctx, database.CreateDeviceParams{
		ProjectID:  project.ID,
		DeviceID:   req.DeviceID,
		DeviceName: deviceName,
	})
	if err != nil {
		log.Printf("Failed to create device: %v", err)
		return c.JSON(http.StatusInternalServerError, DeviceRegisterResponse{
			Success: false,
			Error:   "Failed to register device",
		})
	}

	return c.JSON(http.StatusOK, DeviceRegisterResponse{
		Success:    true,
		DeviceID:   device.DeviceID,
		CourseName: nil,
		Message:    "Device registered. No course assigned yet.",
	})
}

// POST /api/v1/locations
func (h *LocationHandler) CreateLocations(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. X-Project-Api-Key ヘッダーからAPIキーを取得
	apiKey := c.Request().Header.Get("X-Project-Api-Key")
	if apiKey == "" {
		return c.JSON(http.StatusUnauthorized, LocationResponse{
			Success: false,
			Error:   "API key is required",
		})
	}

	// 2. APIキーでプロジェクトを検索し認証
	project, err := h.DB.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, LocationResponse{
				Success: false,
				Error:   "Invalid API key",
			})
		}
		log.Printf("Database error during API key validation: %v", err)
		return c.JSON(http.StatusInternalServerError, LocationResponse{
			Success: false,
			Error:   "Internal server error during API key validation",
		})
	}

	var req LocationRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, LocationResponse{
			Success: false,
			Error:   "Invalid request format",
		})
	}

	// バリデーション
	if req.DeviceID == "" {
		return c.JSON(http.StatusBadRequest, LocationResponse{
			Success: false,
			Error:   "device_id is required",
		})
	}

	if len(req.Locations) == 0 {
		return c.JSON(http.StatusBadRequest, LocationResponse{
			Success: false,
			Error:   "locations array cannot be empty",
		})
	}

	// 3. device_id からコース名を取得
	device, err := h.DB.GetDeviceByDeviceID(ctx, database.GetDeviceByDeviceIDParams{
		ProjectID: project.ID,
		DeviceID:  req.DeviceID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, LocationResponse{
				Success: false,
				Error:   "Device not found. Register device first using POST /api/v1/devices",
			})
		}
		log.Printf("Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, LocationResponse{
			Success: false,
			Error:   "Failed to retrieve device",
		})
	}

	if !device.CourseName.Valid || device.CourseName.String == "" {
		return c.JSON(http.StatusBadRequest, LocationResponse{
			Success: false,
			Error:   "No course assigned to this device",
		})
	}
	courseName := device.CourseName.String

	// 最終通信日時を更新
	_ = h.DB.UpdateDeviceLastSeen(ctx, database.UpdateDeviceLastSeenParams{
		ProjectID: project.ID,
		DeviceID:  req.DeviceID,
	})

	// 各位置情報を保存
	recorded := 0
	for _, loc := range req.Locations {
		// タイムスタンプのパース
		timestamp, err := time.Parse(time.RFC3339, loc.Timestamp)
		if err != nil {
			log.Printf("Invalid timestamp format: %s, error: %v", loc.Timestamp, err)
			continue // スキップして次へ
		}

		// sql.Null型への変換
		var accuracy, speed, bearing sql.NullFloat64
		var batteryLevel sql.NullInt64

		if loc.Accuracy != nil {
			accuracy = sql.NullFloat64{Float64: *loc.Accuracy, Valid: true}
		}
		if loc.Speed != nil {
			speed = sql.NullFloat64{Float64: *loc.Speed, Valid: true}
		}
		if loc.Bearing != nil {
			bearing = sql.NullFloat64{Float64: *loc.Bearing, Valid: true}
		}
		if loc.BatteryLevel != nil {
			batteryLevel = sql.NullInt64{Int64: *loc.BatteryLevel, Valid: true}
		}

		// データベースに挿入
		err = h.DB.CreateLocationLog(ctx, database.CreateLocationLogParams{
			ProjectID:    project.ID,
			CourseName:   courseName,
			DeviceID:     sql.NullString{String: req.DeviceID, Valid: true},
			Latitude:     loc.Latitude,
			Longitude:    loc.Longitude,
			Timestamp:    timestamp,
			Accuracy:     accuracy,
			Speed:        speed,
			Bearing:      bearing,
			BatteryLevel: batteryLevel,
		})

		if err != nil {
			log.Printf("Failed to insert location log: %v", err)
			continue // エラーがあっても次へ
		}

		recorded++
	}

	// レスポンス
	if recorded == 0 {
		return c.JSON(http.StatusBadRequest, LocationResponse{
			Success: false,
			Error:   "No valid locations were recorded",
		})
	}

	return c.JSON(http.StatusOK, LocationResponse{
		Success:  true,
		Recorded: recorded,
		Message:  fmt.Sprintf("%d locations recorded", recorded),
	})
}

// 写真メタデータAPI用の構造体
type PhotoMetadataRequest struct {
	DeviceID      string  `json:"device_id"`
	DevicePhotoID string  `json:"device_photo_id"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	TakenAt       string  `json:"taken_at"`
}

type RouteStopInfo struct {
	ID             int64   `json:"id"`
	Sequence       string  `json:"sequence"`
	StopName       string  `json:"stop_name"`
	Address        string  `json:"address,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	DistanceMeters float64 `json:"distance_meters"`
}

type PhotoMetadataResponse struct {
	Success     bool           `json:"success"`
	PhotoID     int64          `json:"photo_id,omitempty"`
	MatchedStop *RouteStopInfo `json:"matched_stop,omitempty"`
	Message     string         `json:"message,omitempty"`
	Error       string         `json:"error,omitempty"`
}

// POST /api/v1/photos
func (h *LocationHandler) CreatePhotoMetadata(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. X-Project-Api-Key ヘッダーからAPIキーを取得
	apiKey := c.Request().Header.Get("X-Project-Api-Key")
	if apiKey == "" {
		return c.JSON(http.StatusUnauthorized, PhotoMetadataResponse{
			Success: false,
			Error:   "API key is required",
		})
	}

	// 2. APIキーでプロジェクトを検索し認証
	project, err := h.DB.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, PhotoMetadataResponse{
				Success: false,
				Error:   "Invalid API key",
			})
		}
		log.Printf("Database error during API key validation: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoMetadataResponse{
			Success: false,
			Error:   "Internal server error during API key validation",
		})
	}

	var req PhotoMetadataRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, PhotoMetadataResponse{
			Success: false,
			Error:   "Invalid request format",
		})
	}

	// バリデーション
	if req.DeviceID == "" {
		return c.JSON(http.StatusBadRequest, PhotoMetadataResponse{
			Success: false,
			Error:   "device_id is required",
		})
	}
	if req.DevicePhotoID == "" {
		return c.JSON(http.StatusBadRequest, PhotoMetadataResponse{
			Success: false,
			Error:   "device_photo_id is required",
		})
	}

	// 3. device_id からコース名を取得
	device, err := h.DB.GetDeviceByDeviceID(ctx, database.GetDeviceByDeviceIDParams{
		ProjectID: project.ID,
		DeviceID:  req.DeviceID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, PhotoMetadataResponse{
				Success: false,
				Error:   "Device not found. Register device first using POST /api/v1/devices",
			})
		}
		log.Printf("Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoMetadataResponse{
			Success: false,
			Error:   "Failed to retrieve device",
		})
	}

	if !device.CourseName.Valid || device.CourseName.String == "" {
		return c.JSON(http.StatusBadRequest, PhotoMetadataResponse{
			Success: false,
			Error:   "No course assigned to this device",
		})
	}
	courseName := device.CourseName.String

	// 最終通信日時を更新
	_ = h.DB.UpdateDeviceLastSeen(ctx, database.UpdateDeviceLastSeenParams{
		ProjectID: project.ID,
		DeviceID:  req.DeviceID,
	})

	// タイムスタンプのパース
	takenAt, err := time.Parse(time.RFC3339, req.TakenAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, PhotoMetadataResponse{
			Success: false,
			Error:   "Invalid taken_at format. Use RFC3339 format (e.g., 2025-12-02T15:04:05+09:00)",
		})
	}

	// 該当コースの停車地一覧を取得
	stops, err := h.DB.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
		ProjectID:  project.ID,
		CourseName: courseName,
	})
	if err != nil {
		log.Printf("Failed to get route stops: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoMetadataResponse{
			Success: false,
			Error:   "Failed to retrieve route stops",
		})
	}

	// 写真の位置から最も近い停車地を探す
	var matchedStop *RouteStopInfo
	var matchedStopID sql.NullInt64
	thresholdMeters := float64(project.ArrivalThresholdMeters.Int64)

	for _, stop := range stops {
		if !stop.Latitude.Valid || !stop.Longitude.Valid {
			continue
		}

		distanceKm := geo.Haversine(
			req.Latitude, req.Longitude,
			stop.Latitude.Float64, stop.Longitude.Float64,
		)
		distance := distanceKm * 1000 // メートルに変換

		// 閾値内で最も近い地点を探す
		if distance <= thresholdMeters {
			if matchedStop == nil || distance < matchedStop.DistanceMeters {
				addr := ""
				if stop.Address.Valid {
					addr = stop.Address.String
				}
				matchedStop = &RouteStopInfo{
					ID:             stop.ID,
					Sequence:       stop.Sequence,
					StopName:       stop.StopName,
					Address:        addr,
					Latitude:       stop.Latitude.Float64,
					Longitude:      stop.Longitude.Float64,
					DistanceMeters: distance,
				}
				matchedStopID = sql.NullInt64{Int64: stop.ID, Valid: true}
			}
		}
	}

	// 写真メタデータをDBに保存
	photo, err := h.DB.CreatePhotoMetadata(ctx, database.CreatePhotoMetadataParams{
		ProjectID:     project.ID,
		CourseName:    courseName,
		DevicePhotoID: req.DevicePhotoID,
		Latitude:      req.Latitude,
		Longitude:     req.Longitude,
		RouteStopID:   matchedStopID,
		TakenAt:       takenAt,
	})
	if err != nil {
		log.Printf("Failed to create photo metadata: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoMetadataResponse{
			Success: false,
			Error:   "Failed to save photo metadata",
		})
	}

	// レスポンス
	response := PhotoMetadataResponse{
		Success:     true,
		PhotoID:     photo.ID,
		MatchedStop: matchedStop,
	}

	if matchedStop != nil {
		response.Message = fmt.Sprintf("Photo registered and matched to stop: %s", matchedStop.StopName)
	} else {
		response.Message = "Photo registered but no matching stop found within threshold"
	}

	return c.JSON(http.StatusOK, response)
}

// 写真アップロードAPI用の構造体
type PhotoUploadResponse struct {
	Success  bool   `json:"success"`
	PhotoID  int64  `json:"photo_id,omitempty"`
	FilePath string `json:"file_path,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// POST /api/v1/photos/upload
func (h *LocationHandler) UploadPhoto(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. X-Project-Api-Key ヘッダーからAPIキーを取得
	apiKey := c.Request().Header.Get("X-Project-Api-Key")
	if apiKey == "" {
		return c.JSON(http.StatusUnauthorized, PhotoUploadResponse{
			Success: false,
			Error:   "API key is required",
		})
	}

	// 2. APIキーでプロジェクトを検索し認証
	project, err := h.DB.GetProjectByAPIKey(ctx, apiKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, PhotoUploadResponse{
				Success: false,
				Error:   "Invalid API key",
			})
		}
		log.Printf("Database error during API key validation: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoUploadResponse{
			Success: false,
			Error:   "Internal server error during API key validation",
		})
	}

	// 3. フォームデータを取得
	devicePhotoID := c.FormValue("device_photo_id")
	if devicePhotoID == "" {
		return c.JSON(http.StatusBadRequest, PhotoUploadResponse{
			Success: false,
			Error:   "device_photo_id is required",
		})
	}

	// 4. 事前登録されたメタデータを取得
	photoMeta, err := h.DB.GetPhotoMetadataByDeviceID(ctx, database.GetPhotoMetadataByDeviceIDParams{
		ProjectID:     project.ID,
		DevicePhotoID: devicePhotoID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, PhotoUploadResponse{
				Success: false,
				Error:   "Photo metadata not found. Register metadata first using POST /api/v1/photos",
			})
		}
		log.Printf("Database error: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoUploadResponse{
			Success: false,
			Error:   "Failed to retrieve photo metadata",
		})
	}

	// 5. 既にアップロード済みかチェック
	if photoMeta.PhotoSynced.Valid && photoMeta.PhotoSynced.Int64 == 1 {
		return c.JSON(http.StatusConflict, PhotoUploadResponse{
			Success: false,
			Error:   "Photo already uploaded",
		})
	}

	// 6. ファイルを取得
	file, err := c.FormFile("photo")
	if err != nil {
		return c.JSON(http.StatusBadRequest, PhotoUploadResponse{
			Success: false,
			Error:   "photo file is required",
		})
	}

	// 7. ファイル形式チェック（JPEG/PNG）
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return c.JSON(http.StatusBadRequest, PhotoUploadResponse{
			Success: false,
			Error:   "Unsupported file format. Use JPEG or PNG",
		})
	}

	// 8. 保存先ディレクトリを作成
	// data/photos/{project_id}/{course_name}/
	saveDir := filepath.Join("data", "photos", fmt.Sprintf("%d", project.ID), photoMeta.CourseName)
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Printf("Failed to create directory: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoUploadResponse{
			Success: false,
			Error:   "Failed to create storage directory",
		})
	}

	// 9. ファイルを保存
	// ファイル名: {device_photo_id}{ext}
	fileName := devicePhotoID + ext
	savePath := filepath.Join(saveDir, fileName)

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, PhotoUploadResponse{
			Success: false,
			Error:   "Failed to open uploaded file",
		})
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		log.Printf("Failed to create file: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoUploadResponse{
			Success: false,
			Error:   "Failed to save file",
		})
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		log.Printf("Failed to write file: %v", err)
		return c.JSON(http.StatusInternalServerError, PhotoUploadResponse{
			Success: false,
			Error:   "Failed to write file",
		})
	}

	// 10. photo_syncedフラグを更新
	if err := h.DB.UpdatePhotoSynced(ctx, photoMeta.ID); err != nil {
		log.Printf("Failed to update photo_synced flag: %v", err)
		// ファイルは保存されたのでエラーにはしない
	}

	// 11. レスポンス
	relativePath := filepath.Join("photos", fmt.Sprintf("%d", project.ID), photoMeta.CourseName, fileName)
	return c.JSON(http.StatusOK, PhotoUploadResponse{
		Success:  true,
		PhotoID:  photoMeta.ID,
		FilePath: relativePath,
		Message:  "Photo uploaded successfully",
	})
}
