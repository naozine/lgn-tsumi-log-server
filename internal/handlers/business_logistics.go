package handlers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/appcontext"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/database"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/geo"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/status"
	"github.com/naozine/project_crud_with_auth_tmpl/web/components"
	"github.com/naozine/project_crud_with_auth_tmpl/web/layouts"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// JST は日本標準時 (componentsからも参照されるためここにも定義)
var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

type LogisticsProjectHandler struct {
	DB *database.Queries
}

func NewLogisticsProjectHandler(db *database.Queries) *LogisticsProjectHandler {
	return &LogisticsProjectHandler{DB: db}
}

// checkPermission は現在のユーザーが書き込み権限を持っているかチェック
func (h *LogisticsProjectHandler) checkPermission(c echo.Context) error {
	role := appcontext.GetUserRole(c.Request().Context())
	if role != "admin" && role != "editor" {
		return echo.NewHTTPError(http.StatusForbidden, "アクセス拒否: 書き込み権限が必要です")
	}
	return nil
}

// ListLogisticsProjects は物流案件一覧ページを表示
func (h *LogisticsProjectHandler) ListLogisticsProjects(c echo.Context) error {
	ctx := c.Request().Context()
	logisticsProjects, err := h.DB.ListLogisticsProjects(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	content := components.LogisticsProjectList(logisticsProjects)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base("物流案件一覧", content).Render(ctx, c.Response().Writer)
}

// NewLogisticsProjectPage は新規物流案件作成ページを表示
func (h *LogisticsProjectHandler) NewLogisticsProjectPage(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}
	ctx := c.Request().Context()
	content := components.LogisticsProjectForm()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base("新規物流案件作成", content).Render(ctx, c.Response().Writer)
}

// CreateLogisticsProject は新規物流案件を作成
func (h *LogisticsProjectHandler) CreateLogisticsProject(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}
	ctx := c.Request().Context()
	name := c.FormValue("name")

	arrivalThreshold := int64(100) // デフォルト値
	if thresholdStr := c.FormValue("arrival_threshold"); thresholdStr != "" {
		if parsed, err := strconv.ParseInt(thresholdStr, 10, 64); err == nil {
			arrivalThreshold = parsed
		}
	}

	_, err := h.DB.CreateLogisticsProject(ctx, database.CreateLogisticsProjectParams{
		Name:                   name,
		ArrivalThresholdMeters: sql.NullInt64{Int64: arrivalThreshold, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	return c.Redirect(http.StatusSeeOther, "/logistics/projects")
}

// ShowLogisticsProject は物流案件詳細ページを表示
func (h *LogisticsProjectHandler) ShowLogisticsProject(c echo.Context) error {
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}

	lp, err := h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	content := components.LogisticsProjectDetail(lp)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base(lp.Name, content).Render(ctx, c.Response().Writer)
}

// EditLogisticsProjectPage は物流案件編集ページを表示
func (h *LogisticsProjectHandler) EditLogisticsProjectPage(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}

	lp, err := h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	content := components.LogisticsProjectEdit(lp)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base("編集: "+lp.Name, content).Render(ctx, c.Response().Writer)
}

// UpdateLogisticsProject は物流案件を更新
func (h *LogisticsProjectHandler) UpdateLogisticsProject(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}

	name := c.FormValue("name")
	arrivalThreshold := int64(100) // デフォルト値
	if thresholdStr := c.FormValue("arrival_threshold"); thresholdStr != "" {
		if parsed, err := strconv.ParseInt(thresholdStr, 10, 64); err == nil {
			arrivalThreshold = parsed
		}
	}

	_, err = h.DB.UpdateLogisticsProject(ctx, database.UpdateLogisticsProjectParams{
		ID:                     lpID,
		Name:                   name,
		ArrivalThresholdMeters: sql.NullInt64{Int64: arrivalThreshold, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/logistics/projects/%d", lpID))
}

// DeleteLogisticsProject は物流案件を削除
func (h *LogisticsProjectHandler) DeleteLogisticsProject(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}

	err = h.DB.DeleteLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusSeeOther, "/logistics/projects")
}

// UploadRoutesPage はCSVアップロードページを表示
func (h *LogisticsProjectHandler) UploadRoutesPage(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}

	lp, err := h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	content := components.RouteUploadForm(lpID, lp.Name)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base("CSVアップロード", content).Render(ctx, c.Response().Writer)
}

// UploadRoutes はCSVファイルをアップロードして既存データを上書き
func (h *LogisticsProjectHandler) UploadRoutes(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}

	// 物流案件の存在確認
	_, err = h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	// ファイルアップロード
	fileHeader, err := c.FormFile("csv_file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "CSVファイルが指定されていません")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ファイルを開けませんでした")
	}
	defer file.Close()

	// CSVパース
	stops, err := parseCP932CSV(file, c.FormValue("has_header") == "true")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("CSVパースエラー: %v", err))
	}

	if len(stops) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "CSVファイルにデータがありません")
	}

	// オプション処理: 「出発」行をスキップ
	if c.FormValue("skip_departure") == "true" {
		stops = filterDepartureRows(stops)
	}

	// オプション処理: 出発時間を調整
	if c.FormValue("adjust_time") == "true" {
		startTime := c.FormValue("start_time")
		if startTime != "" {
			stops = adjustArrivalTimes(stops, startTime)
		}
	}

	if len(stops) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "フィルタ後のデータがありません")
	}

	// 既存データを削除
	err = h.DB.DeleteRouteStopsByProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("既存データ削除失敗: %v", err))
	}

	// 物流案件のCSV情報を更新
	_, err = h.DB.UpdateLogisticsProjectCSV(ctx, database.UpdateLogisticsProjectCSVParams{
		ID: lpID,
		CsvFilename: sql.NullString{
			String: fileHeader.Filename,
			Valid:  true,
		},
		CsvImportedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		CsvRowCount: sql.NullInt64{
			Int64: int64(len(stops)),
			Valid: true,
		},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("物流案件CSV情報更新失敗: %v", err))
	}

	// 停車地データ一括挿入
	for _, stop := range stops {
		err = h.DB.CreateRouteStop(ctx, database.CreateRouteStopParams{
			ProjectID:        lpID,
			CourseName:       stop.CourseName,
			Sequence:         stop.Sequence,
			ArrivalTime:      toNullString(stop.ArrivalTime),
			StopName:         stop.StopName,
			Address:          toNullString(stop.Address),
			Latitude:         toNullFloat64(stop.Latitude),
			Longitude:        toNullFloat64(stop.Longitude),
			StayMinutes:      toNullInt64(stop.StayMinutes),
			WeightKg:         toNullInt64(stop.WeightKg),
			Status:           toNullString(stop.Status),
			PhoneNumber:      toNullString(stop.PhoneNumber),
			Note1:            toNullString(stop.Note1),
			Note2:            toNullString(stop.Note2),
			Note3:            toNullString(stop.Note3),
			DesiredTimeStart: toNullString(stop.DesiredTimeStart),
			DesiredTimeEnd:   toNullString(stop.DesiredTimeEnd),
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("停車地データ挿入失敗: %v", err))
		}
	}

	// PRG: 物流案件詳細ページへリダイレクト
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/logistics/projects/%d", lpID))
}

// ListCourses はコース一覧ページを表示
func (h *LogisticsProjectHandler) ListCourses(c echo.Context) error {
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}

	lp, err := h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	// コース一覧取得（車両名の昇順でソート済み）
	courses, err := h.DB.ListCoursesByProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// 各コースの停車地点数を取得
	courseInfos := make([]components.CourseInfo, 0, len(courses))
	for _, courseName := range courses {
		stops, err := h.DB.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
			ProjectID:  lpID,
			CourseName: courseName,
		})
		if err != nil {
			continue
		}
		courseInfos = append(courseInfos, components.CourseInfo{
			CourseName: courseName,
			StopCount:  len(stops),
		})
	}

	content := components.CourseList(lp, courseInfos)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base("コース一覧", content).Render(ctx, c.Response().Writer)
}

// ShowCourse はコース詳細ページを表示
func (h *LogisticsProjectHandler) ShowCourse(c echo.Context) error {
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}
	courseName := c.Param("course_name")

	lp, err := h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	// コースの停車地取得（順番でソート済み）
	stops, err := h.DB.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
		ProjectID:  lpID,
		CourseName: courseName,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if len(stops) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "コースが見つかりません")
	}

	// 現在位置情報を取得
	var currentLocation *components.CurrentLocationInfo
	latestLog, err := h.DB.GetLatestLocationByCourse(ctx, database.GetLatestLocationByCourseParams{
		ProjectID:  lpID,
		CourseName: courseName,
	})
	if err == nil {
		currentLocation = h.calculateCurrentSection(latestLog, stops)
	}

	content := components.CourseDetail(lp, courseName, stops, currentLocation)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base(courseName, content).Render(ctx, c.Response().Writer)
}

// GetCurrentLocation は現在位置セクションのみを返す（htmx polling用）
func (h *LogisticsProjectHandler) GetCurrentLocation(c echo.Context) error {
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}
	courseName := c.Param("course_name")

	// 物流案件情報を取得（到着判定閾値を使用）
	lp, err := h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	// 到着判定の閾値を取得（デフォルト100m）
	arrivalThresholdM := int64(100)
	if lp.ArrivalThresholdMeters.Valid {
		arrivalThresholdM = lp.ArrivalThresholdMeters.Int64
	}

	// コースの停車地取得
	stops, err := h.DB.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
		ProjectID:  lpID,
		CourseName: courseName,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// 現在位置情報を取得
	var currentLocation *components.CurrentLocationInfo
	latestLog, err := h.DB.GetLatestLocationByCourse(ctx, database.GetLatestLocationByCourseParams{
		ProjectID:  lpID,
		CourseName: courseName,
	})
	if err == nil {
		// 到着判定を行い、必要ならステータスを更新
		h.checkAndUpdateArrival(ctx, latestLog, stops, arrivalThresholdM)

		// 再度停車地を取得（ステータスが更新されている可能性）
		stops, _ = h.DB.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
			ProjectID:  lpID,
			CourseName: courseName,
		})

		currentLocation = h.calculateCurrentSection(latestLog, stops)
	}

	// 部分レンダリング（現在位置セクション＋テーブル）
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return components.CourseLocationStatus(lpID, courseName, stops, currentLocation).Render(ctx, c.Response().Writer)
}

// ShowStop は地点詳細ページを表示
func (h *LogisticsProjectHandler) ShowStop(c echo.Context) error {
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}
	courseName := c.Param("course_name")
	stopID, err := strconv.ParseInt(c.Param("stop_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な地点ID")
	}

	// 物流案件の存在確認
	_, err = h.DB.GetLogisticsProject(ctx, lpID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "物流案件が見つかりません")
	}

	stop, err := h.DB.GetRouteStopByID(ctx, stopID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "地点が見つかりません")
	}

	// 地点がこの案件・コースに属しているか確認
	if stop.ProjectID != lpID || stop.CourseName != courseName {
		return echo.NewHTTPError(http.StatusNotFound, "地点が見つかりません")
	}

	// トラック状況を計算
	truckStatus := h.calculateTruckStatus(ctx, lpID, courseName, stop)

	content := components.StopDetail(lpID, courseName, stopID, stop, truckStatus)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	if c.Request().Header.Get("HX-Request") == "true" {
		return content.Render(ctx, c.Response().Writer)
	}
	return layouts.Base(stop.StopName, content).Render(ctx, c.Response().Writer)
}

// GetStopTruckStatus はトラック状況セクションのみを返す（htmx polling用）
func (h *LogisticsProjectHandler) GetStopTruckStatus(c echo.Context) error {
	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}
	courseName := c.Param("course_name")
	stopID, err := strconv.ParseInt(c.Param("stop_id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な地点ID")
	}

	stop, err := h.DB.GetRouteStopByID(ctx, stopID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "地点が見つかりません")
	}

	// 地点がこの案件・コースに属しているか確認
	if stop.ProjectID != lpID || stop.CourseName != courseName {
		return echo.NewHTTPError(http.StatusNotFound, "地点が見つかりません")
	}

	// トラック状況を計算
	truckStatus := h.calculateTruckStatus(ctx, lpID, courseName, stop)

	// 部分レンダリング（トラック状況セクションのみ）
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return components.StopTruckStatus(truckStatus).Render(ctx, c.Response().Writer)
}

// calculateTruckStatus はトラックの状況を計算する
func (h *LogisticsProjectHandler) calculateTruckStatus(ctx context.Context, lpID int64, courseName string, targetStop database.RouteStop) *components.TruckStatusInfo {
	truckStatus := &components.TruckStatusInfo{
		HasLocation: false,
	}

	// 位置情報があるか確認
	_, err := h.DB.GetLatestLocationByCourse(ctx, database.GetLatestLocationByCourseParams{
		ProjectID:  lpID,
		CourseName: courseName,
	})
	if err != nil {
		return truckStatus
	}
	truckStatus.HasLocation = true

	// コースの全地点を取得
	stops, err := h.DB.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
		ProjectID:  lpID,
		CourseName: courseName,
	})
	if err != nil {
		return truckStatus
	}

	// 対象地点のインデックスを探す
	targetIdx := -1
	for i, s := range stops {
		if s.ID == targetStop.ID {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return truckStatus
	}

	// 最後に到着した地点を探す
	lastArrivedIdx := -1
	for i, s := range stops {
		if s.Status.String == status.Arrived {
			lastArrivedIdx = i
		}
	}

	// 何地点前にいるかを計算
	if lastArrivedIdx == -1 {
		// まだどこにも到着していない → 全地点分前にいる
		truckStatus.StopsAway = targetIdx + 1
		truckStatus.CurrentStopName = "出発前"
	} else {
		// lastArrivedIdx が targetIdx より前なら、その差が「何地点前」
		// lastArrivedIdx が targetIdx と同じなら、この地点に到着済み
		// lastArrivedIdx が targetIdx より後なら、この地点は通過済み
		truckStatus.StopsAway = targetIdx - lastArrivedIdx
		truckStatus.CurrentStopName = stops[lastArrivedIdx].StopName
		truckStatus.LastArrivedStop = stops[lastArrivedIdx].StopName
		truckStatus.ScheduledTime = stops[lastArrivedIdx].ArrivalTime.String

		// 実績時刻を設定
		if stops[lastArrivedIdx].ActualArrivalTime.Valid {
			truckStatus.LastArrivedTime = stops[lastArrivedIdx].ActualArrivalTime.String
		}
		if stops[lastArrivedIdx].ActualDepartureTime.Valid {
			truckStatus.LastDepartedTime = stops[lastArrivedIdx].ActualDepartureTime.String
		}

		// 遅延計算: 予定時刻と実績到着時刻の差
		if stops[lastArrivedIdx].ArrivalTime.Valid && stops[lastArrivedIdx].ActualArrivalTime.Valid {
			scheduledMinutes := parseTimeToMinutesOrZero(stops[lastArrivedIdx].ArrivalTime.String)
			actualMinutes := parseTimeToMinutesOrZero(stops[lastArrivedIdx].ActualArrivalTime.String)
			truckStatus.DelayMinutes = actualMinutes - scheduledMinutes
		}
	}

	return truckStatus
}

// parseTimeToMinutesOrZero は "HH:MM" 形式の時間を分数に変換（エラー時は0）
func parseTimeToMinutesOrZero(timeStr string) int {
	minutes, err := parseTimeToMinutes(timeStr)
	if err != nil {
		return 0
	}
	return minutes
}

// ResetCourseStatus はコースの全地点のステータスを「未訪問」にリセットする
func (h *LogisticsProjectHandler) ResetCourseStatus(c echo.Context) error {
	if err := h.checkPermission(c); err != nil {
		return err
	}

	ctx := c.Request().Context()
	lpID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "無効な案件ID")
	}
	courseName := c.Param("course_name")

	// ステータスをリセット
	err = h.DB.ResetRouteStopsStatusByCourse(ctx, database.ResetRouteStopsStatusByCourseParams{
		Status:     sql.NullString{String: status.Unvisited, Valid: true},
		ProjectID:  lpID,
		CourseName: courseName,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("リセット失敗: %v", err))
	}

	// PRG: コース詳細ページへリダイレクト
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/logistics/projects/%d/courses/%s", lpID, courseName))
}

// checkAndUpdateArrivalDeparture は現在位置から到着・出発判定を行い、ステータスと実績時刻を更新する
// arrivalThresholdM は到着判定の閾値（メートル単位）
//
// 判定ロジック:
// 1. 未到着の地点が範囲内 → 到着（status=到着済, actual_arrival_time=ログ時刻）
// 2. 到着済みで出発時刻なしの地点が範囲外 → 出発（actual_departure_time=ログ時刻）
// 3. 到着済みで出発時刻ありの地点が範囲内 → 出発取り消し（actual_departure_time=NULL）
func (h *LogisticsProjectHandler) checkAndUpdateArrivalDeparture(ctx context.Context, loc database.LocationLog, stops []database.RouteStop, arrivalThresholdM int64) {
	// メートルからキロメートルに変換
	arrivalThresholdKm := float64(arrivalThresholdM) / 1000.0
	// JSTに変換して時刻文字列を生成
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	logTimeStr := loc.Timestamp.In(jst).Format("15:04")

	for _, stop := range stops {
		// 座標がない地点はスキップ
		if !stop.Latitude.Valid || !stop.Longitude.Valid {
			continue
		}

		// 距離を計算
		dist := geo.Haversine(loc.Latitude, loc.Longitude, stop.Latitude.Float64, stop.Longitude.Float64)
		isWithinRange := dist < arrivalThresholdKm

		if stop.Status.String != status.Arrived {
			// 未到着の地点
			if isWithinRange {
				// 範囲内に入った → 到着（出発時刻はクリア）
				_ = h.DB.UpdateRouteStopArrival(ctx, database.UpdateRouteStopArrivalParams{
					Status:            sql.NullString{String: status.Arrived, Valid: true},
					ActualArrivalTime: sql.NullString{String: logTimeStr, Valid: true},
					ID:                stop.ID,
				})
			}
		} else {
			// 到着済みの地点
			if !stop.ActualDepartureTime.Valid || stop.ActualDepartureTime.String == "" {
				// 出発時刻がまだない
				if !isWithinRange {
					// 範囲外に出た → 出発
					_ = h.DB.UpdateRouteStopDeparture(ctx, database.UpdateRouteStopDepartureParams{
						ActualDepartureTime: sql.NullString{String: logTimeStr, Valid: true},
						ID:                  stop.ID,
					})
				}
			} else {
				// 出発時刻がある
				if isWithinRange {
					// 範囲内に戻ってきた → 出発取り消し
					_ = h.DB.ClearRouteStopDeparture(ctx, stop.ID)
				}
			}
		}
	}
}

// checkAndUpdateArrival は後方互換性のためのラッパー（非推奨）
func (h *LogisticsProjectHandler) checkAndUpdateArrival(ctx context.Context, loc database.LocationLog, stops []database.RouteStop, arrivalThresholdM int64) {
	h.checkAndUpdateArrivalDeparture(ctx, loc, stops, arrivalThresholdM)
}

// calculateCurrentSection は現在位置から走行中の区間を計算する
func (h *LogisticsProjectHandler) calculateCurrentSection(loc database.LocationLog, stops []database.RouteStop) *components.CurrentLocationInfo {
	if len(stops) == 0 {
		return nil
	}

	info := &components.CurrentLocationInfo{
		Latitude:  loc.Latitude,
		Longitude: loc.Longitude,
		Timestamp: loc.Timestamp,
	}

	// 速度を設定（m/s）
	if loc.Speed.Valid {
		info.SpeedMps = loc.Speed.Float64
	}

	// 到着済みの最後の地点を探す（出発地点）
	lastArrivedIdx := -1
	for i, stop := range stops {
		if stop.Status.String == status.Arrived {
			lastArrivedIdx = i
		}
	}

	// 出発地点を設定
	if lastArrivedIdx >= 0 {
		info.FromStop = stops[lastArrivedIdx].StopName
		info.FromStopIdx = lastArrivedIdx
	}

	// 次の未到着地点を探す（目的地点）
	nextStopIdx := -1
	for i, stop := range stops {
		if stop.Status.String != status.Arrived {
			nextStopIdx = i
			break
		}
	}

	// 目的地点を設定
	if nextStopIdx >= 0 {
		nextStop := stops[nextStopIdx]
		info.ToStop = nextStop.StopName
		info.ToStopIdx = nextStopIdx

		// 目的地までの距離を計算
		if nextStop.Latitude.Valid && nextStop.Longitude.Valid {
			info.ToDistanceKm = geo.Haversine(
				loc.Latitude, loc.Longitude,
				nextStop.Latitude.Float64, nextStop.Longitude.Float64,
			)

			// ETA計算（速度が有効な場合）
			if info.SpeedMps > 0 {
				distanceM := info.ToDistanceKm * 1000
				etaSeconds := distanceM / info.SpeedMps
				info.ToEtaMinutes = etaSeconds / 60
			}
		}
	}

	// 後方互換性のため、最寄り地点情報も設定
	info.NearestStop = info.ToStop
	info.NearestStopIdx = info.ToStopIdx
	info.DistanceKm = info.ToDistanceKm
	info.NextStop = ""
	info.NextStopIdx = -1
	info.NextDistanceKm = 0

	return info
}

// RouteStop はCSVの1行分のデータ
type RouteStop struct {
	CourseName       string
	Sequence         string
	ArrivalTime      string
	StopName         string
	Address          string
	Latitude         float64
	Longitude        float64
	StayMinutes      int64
	WeightKg         int64
	Status           string
	PhoneNumber      string
	Note1            string
	Note2            string
	Note3            string
	DesiredTimeStart string
	DesiredTimeEnd   string
}

// parseCP932CSV はCP932エンコードされたCSVファイルをパース
func parseCP932CSV(file multipart.File, hasHeader bool) ([]RouteStop, error) {
	// CP932 -> UTF-8変換
	reader := transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	csvReader := csv.NewReader(reader)

	// フィールド数のチェックを無効化（行末のカンマ対応）
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true

	// ヘッダースキップ
	if hasHeader {
		_, err := csvReader.Read()
		if err != nil {
			return nil, fmt.Errorf("ヘッダー読み込みエラー: %w", err)
		}
	}

	var stops []RouteStop
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV読み込みエラー: %w", err)
		}

		// 17カラム必要（CSVの構造に基づく）
		if len(record) < 17 {
			continue // スキップ
		}

		latitude, _ := strconv.ParseFloat(record[5], 64)
		longitude, _ := strconv.ParseFloat(record[6], 64)
		stayMinutes, _ := strconv.ParseInt(record[7], 10, 64)
		weightKg, _ := strconv.ParseInt(record[8], 10, 64)

		stop := RouteStop{
			CourseName:       record[0],
			Sequence:         record[1],
			ArrivalTime:      record[2],
			StopName:         record[3],
			Address:          record[4],
			Latitude:         latitude,
			Longitude:        longitude,
			StayMinutes:      stayMinutes,
			WeightKg:         weightKg,
			Status:           record[9],
			PhoneNumber:      record[11],
			Note1:            record[12],
			Note2:            record[13],
			Note3:            record[14],
			DesiredTimeStart: record[15],
			DesiredTimeEnd:   record[16],
		}
		stops = append(stops, stop)
	}

	return stops, nil
}

// ヘルパー関数: stringをsql.NullStringに変換
func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// ヘルパー関数: float64をsql.NullFloat64に変換
func toNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: f != 0}
}

// ヘルパー関数: int64をsql.NullInt64に変換
func toNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: i != 0}
}

// parseTimeToMinutes は "HH:MM" 形式の時間を分数に変換する
func parseTimeToMinutes(timeStr string) (int, error) {
	if timeStr == "" {
		return 0, fmt.Errorf("空の時間文字列")
	}
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("無効な時間形式: %s", timeStr)
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	return hours*60 + minutes, nil
}

// minutesToTime は分数を "HH:MM" 形式に変換する
func minutesToTime(minutes int) string {
	// 負の値や24時間超えに対応
	for minutes < 0 {
		minutes += 24 * 60
	}
	minutes = minutes % (24 * 60)
	return fmt.Sprintf("%02d:%02d", minutes/60, minutes%60)
}

// filterDepartureRows は「出発」行を除外する
func filterDepartureRows(stops []RouteStop) []RouteStop {
	var filtered []RouteStop
	for _, stop := range stops {
		if stop.Sequence != "出発" {
			filtered = append(filtered, stop)
		}
	}
	return filtered
}

// adjustArrivalTimes は各コースの到着時間を調整する
func adjustArrivalTimes(stops []RouteStop, newStartTime string) []RouteStop {
	if len(stops) == 0 {
		return stops
	}

	newStartMinutes, err := parseTimeToMinutes(newStartTime)
	if err != nil {
		return stops // パースエラーの場合は変更しない
	}

	// コースごとにグループ化
	courseStops := make(map[string][]int) // コース名 -> stopsのインデックス
	for i, stop := range stops {
		courseStops[stop.CourseName] = append(courseStops[stop.CourseName], i)
	}

	// 各コースの時間を調整
	for _, indices := range courseStops {
		if len(indices) == 0 {
			continue
		}

		// 先頭地点の元の到着時間を取得
		firstIdx := indices[0]
		originalStartMinutes, err := parseTimeToMinutes(stops[firstIdx].ArrivalTime)
		if err != nil {
			continue // パースエラーの場合はこのコースをスキップ
		}

		// 差分を計算
		diff := newStartMinutes - originalStartMinutes

		// 全地点の到着時間をシフト
		for _, idx := range indices {
			originalMinutes, err := parseTimeToMinutes(stops[idx].ArrivalTime)
			if err != nil {
				continue
			}
			stops[idx].ArrivalTime = minutesToTime(originalMinutes + diff)
		}
	}

	return stops
}
