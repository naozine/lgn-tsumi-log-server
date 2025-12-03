package mdm

import (
	"sync"
	"time"
)

// DeviceCache はデバイス一覧のキャッシュを管理する
type DeviceCache struct {
	devices   []Device
	fetchedAt time.Time
	ttl       time.Duration
	mu        sync.RWMutex
}

// NewDeviceCache は新しいキャッシュを作成する
func NewDeviceCache(ttlSeconds int) *DeviceCache {
	if ttlSeconds <= 0 {
		ttlSeconds = 300 // デフォルト5分
	}
	return &DeviceCache{
		ttl: time.Duration(ttlSeconds) * time.Second,
	}
}

// Get はキャッシュされたデバイス一覧を返す
// キャッシュが有効な場合は (devices, true) を返す
// キャッシュが無効または期限切れの場合は (nil, false) を返す
func (c *DeviceCache) Get() ([]Device, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.devices == nil {
		return nil, false
	}

	if time.Since(c.fetchedAt) > c.ttl {
		return nil, false
	}

	return c.devices, true
}

// Set はデバイス一覧をキャッシュに保存する
func (c *DeviceCache) Set(devices []Device) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.devices = devices
	c.fetchedAt = time.Now()
}

// Clear はキャッシュをクリアする
func (c *DeviceCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.devices = nil
	c.fetchedAt = time.Time{}
}

// FetchedAt はキャッシュの取得時刻を返す
func (c *DeviceCache) FetchedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fetchedAt
}

// TTL はキャッシュの有効期間を返す
func (c *DeviceCache) TTL() time.Duration {
	return c.ttl
}
