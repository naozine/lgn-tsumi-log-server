package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/database"
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
	CourseName string         `json:"course_name"`
	Locations  []LocationData `json:"locations"`
}

// レスポンス用の構造体
type LocationResponse struct {
	Success  bool   `json:"success"`
	Recorded int    `json:"recorded,omitempty"`
	Message  string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
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
	// req.ProjectIDはAPIキーで認証されたものを使用するため、クライアントからの値は無視し、バリデーションも不要
	if req.CourseName == "" {
		return c.JSON(http.StatusBadRequest, LocationResponse{
			Success: false,
			Error:   "course_name is required",
		})
	}

	if len(req.Locations) == 0 {
		return c.JSON(http.StatusBadRequest, LocationResponse{
			Success: false,
			Error:   "locations array cannot be empty",
		})
	}

	// 物流案件の存在確認 (APIキーで既に認証済みのため、既存のGetProjectの呼び出しは不要)

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
			ProjectID:    project.ID, // ここをAPIキーで認証されたproject.IDに変更
			CourseName:   req.CourseName,
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
