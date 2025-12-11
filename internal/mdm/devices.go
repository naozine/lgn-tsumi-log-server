package mdm

import (
	"fmt"
	"strconv"
)

// Device はMDM管理下のデバイス情報
type Device struct {
	DeviceID          int64     `json:"device_id,string"`
	DeviceName        string    `json:"device_name"`
	PlatformType      string    `json:"platform_type"` // "1"=iOS, "2"=Android, "3"=Windows（詳細APIは数値文字列）
	OSVersion         string    `json:"os_version"`
	OSName            string    `json:"os_name"` // 詳細APIのみ（例: "VANILLA_ICE_CREAM"）
	UDID              string    `json:"udid"`
	SerialNumber      string    `json:"serial_number"`
	IMEI              any       `json:"imei"` // 一覧API: []string, 詳細API: string
	Model             string    `json:"model"`
	ModelName         string    `json:"model_name"`   // 詳細APIのみ
	Manufacturer      string    `json:"manufacturer"` // 詳細APIのみ
	User              *User     `json:"user,omitempty"`
	BatteryLevel      any       `json:"battery_level"`         // 一覧: int文字列, 詳細: float文字列
	ManagedStatus     int       `json:"managed_status,string"` // 2=Managed
	Security          *Security `json:"security,omitempty"`
	Summary           *Summary  `json:"summary,omitempty"`
	Sims              []Sim     `json:"sims,omitempty"`       // 詳細APIのみ、SIM情報配列
	Network           *Network  `json:"network,omitempty"`    // 詳細APIのみ
	IsSupervised      bool      `json:"is_supervised"`        // 詳細APIのみ
	IsLostModeEnabled bool      `json:"is_lost_mode_enabled"` // 詳細APIのみ
}

// Sim はSIM情報
type Sim struct {
	SimID                 int64  `json:"sim_id,string"`
	IMEI                  string `json:"imei"`
	ICCID                 string `json:"iccid"`
	PhoneNumber           string `json:"phone_number"`
	Slot                  string `json:"slot"`
	CurrentCarrierNetwork string `json:"current_carrier_network"`
}

// Network はネットワーク情報
type Network struct {
	WifiIP                string `json:"wifi_ip"`
	WifiMAC               string `json:"wifi_mac"`
	BluetoothMAC          string `json:"bluetooth_mac"`
	CurrentCarrierNetwork string `json:"current_carrier_network"`
}

// Security はデバイスのセキュリティ情報
type Security struct {
	PasscodePresent   bool `json:"passcode_present"`   // 詳細API
	Passcode          bool `json:"passcode"`           // 一覧API（互換性のため残す）
	StorageEncryption bool `json:"storage_encryption"` // 詳細API
	Encryption        bool `json:"device_encryption"`  // 一覧API（互換性のため残す）
	DeviceRooted      bool `json:"device_rooted"`      // 詳細API
	Jailbroken        bool `json:"is_device_rooted"`   // 一覧API（互換性のため残す）
	LostModeEnabled   bool `json:"lost_mode_enabled"`
	PasscodeComplaint bool `json:"passcode_complaint"` // 詳細APIのみ
}

// Summary はデバイスの概要情報
type Summary struct {
	ProfileCount string `json:"profile_count"` // 文字列で返ってくる
	AppCount     string `json:"app_count"`
	DocCount     string `json:"doc_count"`
	GroupCount   string `json:"group_count"`
}

// User はデバイスに紐づくユーザー情報
type User struct {
	UserID    int64  `json:"user_id,string"`
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

// DevicesResponse はデバイス一覧APIのレスポンス
type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

// PlatformIcon はプラットフォームに応じたアイコンを返す
func (d *Device) PlatformIcon() string {
	switch d.PlatformType {
	case "ios", "1":
		return "🍎"
	case "android", "2":
		return "🤖"
	case "windows", "3":
		return "💻"
	default:
		return "📱"
	}
}

// PlatformDisplayName はプラットフォームの表示名を返す
func (d *Device) PlatformDisplayName() string {
	switch d.PlatformType {
	case "ios", "1":
		return "iOS"
	case "android", "2":
		return "Android"
	case "windows", "3":
		return "Windows"
	default:
		return d.PlatformType
	}
}

// IMEIString はIMEI情報を文字列で返す
// 一覧API: []string, 詳細API: string, Simsからも取得
func (d *Device) IMEIString() string {
	// 詳細APIのSims配列から取得
	if len(d.Sims) > 0 {
		var imeis []string
		for _, sim := range d.Sims {
			if sim.IMEI != "" && sim.IMEI != "--" {
				imeis = append(imeis, sim.IMEI)
			}
		}
		if len(imeis) > 0 {
			result := ""
			for i, imei := range imeis {
				if i > 0 {
					result += ", "
				}
				result += imei
			}
			return result
		}
	}

	// IMEIフィールドから取得（型によって処理を分岐）
	switch v := d.IMEI.(type) {
	case string:
		if v == "" || v == "--" {
			return "-"
		}
		return v
	case []interface{}:
		if len(v) == 0 {
			return "-"
		}
		result := ""
		for i, imei := range v {
			if i > 0 {
				result += ", "
			}
			if s, ok := imei.(string); ok {
				result += s
			}
		}
		return result
	case []string:
		if len(v) == 0 {
			return "-"
		}
		result := ""
		for i, imei := range v {
			if i > 0 {
				result += ", "
			}
			result += imei
		}
		return result
	}
	return "-"
}

// ManagedStatusString は管理ステータスの表示名を返す
func (d *Device) ManagedStatusString() string {
	switch d.ManagedStatus {
	case 2:
		return "管理中"
	case 1:
		return "登録済み"
	default:
		return "不明"
	}
}

// BatteryLevelString はバッテリーレベルを文字列で返す
func (d *Device) BatteryLevelString() string {
	switch v := d.BatteryLevel.(type) {
	case string:
		// "94.0" のような文字列
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return fmt.Sprintf("%.0f%%", f)
		}
		return "-"
	case float64:
		return fmt.Sprintf("%.0f%%", v)
	case int:
		if v <= 0 {
			return "-"
		}
		return fmt.Sprintf("%d%%", v)
	}
	return "-"
}

// HasPasscode はパスコードが設定されているか返す（一覧/詳細API両対応）
func (s *Security) HasPasscode() bool {
	return s.PasscodePresent || s.Passcode
}

// IsEncrypted は暗号化されているか返す（一覧/詳細API両対応）
func (s *Security) IsEncrypted() bool {
	return s.StorageEncryption || s.Encryption
}

// IsRooted はroot/jailbreakされているか返す（一覧/詳細API両対応）
func (s *Security) IsRooted() bool {
	return s.DeviceRooted || s.Jailbroken
}

// GetProfileCount はプロファイル数を返す
func (s *Summary) GetProfileCount() string {
	if s.ProfileCount == "" {
		return "0"
	}
	return s.ProfileCount
}

// GetAppCount はアプリ数を返す
func (s *Summary) GetAppCount() string {
	if s.AppCount == "" {
		return "0"
	}
	return s.AppCount
}

// GetDocCount はドキュメント数を返す
func (s *Summary) GetDocCount() string {
	if s.DocCount == "" {
		return "0"
	}
	return s.DocCount
}

// GetGroupCount はグループ数を返す
func (s *Summary) GetGroupCount() string {
	if s.GroupCount == "" {
		return "0"
	}
	return s.GroupCount
}
