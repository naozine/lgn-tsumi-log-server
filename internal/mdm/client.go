package mdm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client はMDM APIクライアント
type Client struct {
	baseURL    string
	oauth      *ZohoOAuthClient
	cache      *DeviceCache
	httpClient *http.Client
}

// NewClient は新しいMDMクライアントを作成する
func NewClient(baseURL string, oauth *ZohoOAuthClient, cacheTTLSeconds int) *Client {
	return &Client{
		baseURL: baseURL,
		oauth:   oauth,
		cache:   NewDeviceCache(cacheTTLSeconds),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsConfigured はMDM設定が完了しているか確認する
func (c *Client) IsConfigured() bool {
	return c.baseURL != "" && c.oauth != nil && c.oauth.IsConfigured()
}

// ListDevices はMDM管理下のデバイス一覧を取得する
func (c *Client) ListDevices(ctx context.Context) ([]Device, error) {
	// キャッシュをチェック
	if devices, ok := c.cache.Get(); ok {
		return devices, nil
	}

	// APIから取得
	devices, err := c.fetchDevices(ctx)
	if err != nil {
		return nil, err
	}

	// キャッシュに保存
	c.cache.Set(devices)

	return devices, nil
}

// fetchDevices はAPIからデバイス一覧を取得する
func (c *Client) fetchDevices(ctx context.Context) ([]Device, error) {
	// アクセストークンを取得
	token, err := c.oauth.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// リクエストを作成
	endpoint := fmt.Sprintf("%s/api/v1/mdm/devices", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Zoho-oauthtoken %s", token))
	req.Header.Set("Accept", "application/json")

	// リクエストを送信
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// レスポンスをパース
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var devicesResp DevicesResponse
	if err := json.Unmarshal(body, &devicesResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return devicesResp.Devices, nil
}

// ClearCache はキャッシュをクリアする
func (c *Client) ClearCache() {
	c.cache.Clear()
}

// CacheFetchedAt はキャッシュの取得時刻を返す
func (c *Client) CacheFetchedAt() time.Time {
	return c.cache.FetchedAt()
}

// CacheTTL はキャッシュの有効期間を返す
func (c *Client) CacheTTL() time.Duration {
	return c.cache.TTL()
}

// ListApps はMDMに登録されたアプリ一覧を取得する
func (c *Client) ListApps(ctx context.Context) ([]App, error) {
	token, err := c.oauth.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v1/mdm/apps", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Zoho-oauthtoken %s", token))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch apps: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var appsResp AppsResponse
	if err := json.Unmarshal(body, &appsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return appsResp.Apps, nil
}

// DistributeAppToDevices はアプリをデバイスに配布する
func (c *Client) DistributeAppToDevices(ctx context.Context, appID, releaseLabelID int64, deviceIDs []int64, silentInstall bool) error {
	token, err := c.oauth.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v1/mdm/apps/%d/labels/%d/devices", c.baseURL, appID, releaseLabelID)

	reqBody := DistributeAppsRequest{
		DeviceIDs:     deviceIDs,
		SilentInstall: silentInstall,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Zoho-oauthtoken %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to distribute app: %w", err)
	}
	defer resp.Body.Close()

	// 202 Accepted が成功
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// DistributeAppToGroups はアプリをグループに配布する
func (c *Client) DistributeAppToGroups(ctx context.Context, appID, releaseLabelID int64, groupIDs []int64, silentInstall bool) error {
	token, err := c.oauth.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/v1/mdm/apps/%d/labels/%d/groups", c.baseURL, appID, releaseLabelID)

	reqBody := DistributeAppsRequest{
		GroupIDs:      groupIDs,
		SilentInstall: silentInstall,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Zoho-oauthtoken %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to distribute app: %w", err)
	}
	defer resp.Body.Close()

	// 202 Accepted が成功
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
