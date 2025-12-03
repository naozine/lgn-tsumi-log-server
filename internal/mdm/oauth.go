package mdm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ZohoOAuthClient はZoho OAuthの認証を管理するクライアント
type ZohoOAuthClient struct {
	clientID     string
	clientSecret string
	refreshToken string
	accountsURL  string

	accessToken string
	expiresAt   time.Time
	mu          sync.RWMutex
	httpClient  *http.Client
}

// OAuthTokenResponse はトークンエンドポイントのレスポンス
type OAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"` // 秒数（通常3600）
	APIEndpoint string `json:"api_domain"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error,omitempty"`
}

// NewZohoOAuthClient は新しいOAuthクライアントを作成する
func NewZohoOAuthClient(clientID, clientSecret, refreshToken, accountsURL string) *ZohoOAuthClient {
	if accountsURL == "" {
		accountsURL = "https://accounts.zoho.com" // デフォルトはUS
	}
	return &ZohoOAuthClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
		accountsURL:  accountsURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAccessToken は有効なアクセストークンを返す
// 期限切れの場合は自動でリフレッシュする
func (c *ZohoOAuthClient) GetAccessToken() (string, error) {
	c.mu.RLock()
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		token := c.accessToken
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	// トークンをリフレッシュ
	return c.refreshAccessToken()
}

// refreshAccessToken はRefresh Tokenを使って新しいAccess Tokenを取得する
func (c *ZohoOAuthClient) refreshAccessToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// ダブルチェック（他のgoroutineがすでにリフレッシュした可能性）
	if c.accessToken != "" && time.Now().Before(c.expiresAt) {
		return c.accessToken, nil
	}

	// リクエストパラメータを構築
	data := url.Values{}
	data.Set("refresh_token", c.refreshToken)
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "refresh_token")

	// トークンエンドポイントにPOST
	endpoint := fmt.Sprintf("%s/oauth/v2/token", c.accountsURL)
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("oauth error: %s", tokenResp.Error)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	// トークンを保存（有効期限の5分前に期限切れとみなす）
	c.accessToken = tokenResp.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return c.accessToken, nil
}

// IsConfigured はOAuth設定が完了しているか確認する
func (c *ZohoOAuthClient) IsConfigured() bool {
	return c.clientID != "" && c.clientSecret != "" && c.refreshToken != ""
}
