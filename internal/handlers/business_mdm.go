package handlers

import (
	"net/http"
	"strconv"

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
