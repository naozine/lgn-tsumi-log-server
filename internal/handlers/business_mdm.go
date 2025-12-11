package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/mdm"
	"github.com/naozine/project_crud_with_auth_tmpl/web/components"
	"github.com/naozine/project_crud_with_auth_tmpl/web/layouts"
)

// MDMHandler はMDM関連のハンドラ
type MDMHandler struct {
	MDM *mdm.Client
}

// NewMDMHandler は新しいMDMHandlerを作成する
func NewMDMHandler(mdmClient *mdm.Client) *MDMHandler {
	return &MDMHandler{
		MDM: mdmClient,
	}
}

// MDMTop はMDM管理のトップページを表示する
func (h *MDMHandler) MDMTop(c echo.Context) error {
	ctx := c.Request().Context()

	// MDMが設定されていない場合
	if h.MDM == nil || !h.MDM.IsConfigured() {
		content := components.MDMNotConfigured()
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDM管理", content).Render(ctx, c.Response().Writer)
	}

	content := components.MDMTop()
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return layouts.Base("MDM管理", content).Render(ctx, c.Response().Writer)
}

// ListMDMDevices はMDM管理下のデバイス一覧を表示する
func (h *MDMHandler) ListMDMDevices(c echo.Context) error {
	ctx := c.Request().Context()

	// MDMが設定されていない場合
	if h.MDM == nil || !h.MDM.IsConfigured() {
		content := components.MDMNotConfigured()
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDM管理", content).Render(ctx, c.Response().Writer)
	}

	// デバイス一覧を取得
	devices, err := h.MDM.ListDevices(ctx)
	if err != nil {
		content := components.MDMError(err.Error())
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDMデバイス一覧", content).Render(ctx, c.Response().Writer)
	}

	// キャッシュ情報
	cachedAt := h.MDM.CacheFetchedAt()

	content := components.MDMDeviceList(devices, cachedAt)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return layouts.Base("MDMデバイス一覧", content).Render(ctx, c.Response().Writer)
}

// ShowMDMDevice はMDMデバイスの詳細を表示する
func (h *MDMHandler) ShowMDMDevice(c echo.Context) error {
	ctx := c.Request().Context()

	// MDMが設定されていない場合
	if h.MDM == nil || !h.MDM.IsConfigured() {
		content := components.MDMNotConfigured()
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDM管理", content).Render(ctx, c.Response().Writer)
	}

	// デバイスIDを取得
	deviceIDStr := c.Param("id")
	deviceID, err := strconv.ParseInt(deviceIDStr, 10, 64)
	if err != nil {
		content := components.MDMError("無効なデバイスIDです")
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusBadRequest)
		return layouts.Base("MDMデバイス詳細", content).Render(ctx, c.Response().Writer)
	}

	// デバイス詳細を取得
	device, err := h.MDM.GetDevice(ctx, deviceID)
	if err != nil {
		content := components.MDMError(err.Error())
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDMデバイス詳細", content).Render(ctx, c.Response().Writer)
	}

	content := components.MDMDeviceDetail(device)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return layouts.Base("MDMデバイス詳細", content).Render(ctx, c.Response().Writer)
}

// ListMDMApps はMDMに登録されたアプリ一覧を表示する
func (h *MDMHandler) ListMDMApps(c echo.Context) error {
	ctx := c.Request().Context()

	// MDMが設定されていない場合
	if h.MDM == nil || !h.MDM.IsConfigured() {
		content := components.MDMNotConfigured()
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDM管理", content).Render(ctx, c.Response().Writer)
	}

	// アプリ一覧を取得
	apps, err := h.MDM.ListApps(ctx)
	if err != nil {
		content := components.MDMError(err.Error())
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDMアプリ一覧", content).Render(ctx, c.Response().Writer)
	}

	content := components.MDMAppList(apps)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return layouts.Base("MDMアプリ一覧", content).Render(ctx, c.Response().Writer)
}

// APIExplorer はAPI Explorerページを表示する
func (h *MDMHandler) APIExplorer(c echo.Context) error {
	ctx := c.Request().Context()

	// MDMが設定されていない場合
	if h.MDM == nil || !h.MDM.IsConfigured() {
		content := components.MDMNotConfigured()
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return layouts.Base("MDM API Explorer", content).Render(ctx, c.Response().Writer)
	}

	// デバイス一覧を取得（デバイス選択用）
	devices, _ := h.MDM.ListDevices(ctx)

	// 選択中のエンドポイントID
	selectedID := c.QueryParam("endpoint")
	var selectedEndpoint *mdm.APIEndpoint
	if selectedID != "" {
		selectedEndpoint = mdm.GetAPIEndpointByID(selectedID)
	}

	content := components.MDMAPIExplorer(
		mdm.GetAPICategories(),
		mdm.GetEndpointsByCategory(),
		devices,
		selectedEndpoint,
		nil, // 初期表示時はレスポンスなし
		0,
	)
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(http.StatusOK)
	return layouts.Base("MDM API Explorer", content).Render(ctx, c.Response().Writer)
}

// APIExplorerExecute はAPI Explorerからのリクエストを実行する
func (h *MDMHandler) APIExplorerExecute(c echo.Context) error {
	ctx := c.Request().Context()

	// MDMが設定されていない場合
	if h.MDM == nil || !h.MDM.IsConfigured() {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "MDMが設定されていません"})
	}

	// リクエストパラメータを取得
	endpointID := c.FormValue("endpoint_id")
	deviceIDStr := c.FormValue("device_id")
	customParam := c.FormValue("custom_param") // app_id, group_id, user_id など

	endpoint := mdm.GetAPIEndpointByID(endpointID)
	if endpoint == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "無効なエンドポイントです"})
	}

	// パスを構築
	path := endpoint.Path
	if strings.Contains(path, "{device_id}") {
		if deviceIDStr == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "デバイスを選択してください"})
		}
		path = strings.Replace(path, "{device_id}", deviceIDStr, 1)
	}
	if strings.Contains(path, "{app_id}") {
		if customParam == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "アプリIDを入力してください"})
		}
		path = strings.Replace(path, "{app_id}", customParam, 1)
	}
	if strings.Contains(path, "{group_id}") {
		if customParam == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "グループIDを入力してください"})
		}
		path = strings.Replace(path, "{group_id}", customParam, 1)
	}
	if strings.Contains(path, "{user_id}") {
		if customParam == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "ユーザーIDを入力してください"})
		}
		path = strings.Replace(path, "{user_id}", customParam, 1)
	}

	// クエリパラメータを構築
	queryParams := make(map[string]string)
	for _, param := range endpoint.Params {
		if param.Type == "query" {
			value := c.FormValue(param.Name)
			if value == "" && param.Default != "" {
				value = param.Default
			}
			if value != "" {
				queryParams[param.Name] = value
			}
		}
	}

	// APIを呼び出し
	responseBody, statusCode, err := h.MDM.CallAPI(ctx, endpoint.Method, path, queryParams)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// JSONを整形
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, responseBody, "", "  "); err != nil {
		// 整形できない場合はそのまま返す
		prettyJSON.Write(responseBody)
	}

	// レスポンスを返す
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status_code": statusCode,
		"path":        "/api/v1/mdm" + path,
		"method":      endpoint.Method,
		"response":    prettyJSON.String(),
	})
}
