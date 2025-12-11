package mdm

// Device はMDM管理下のデバイス情報
type Device struct {
	DeviceID     int64    `json:"device_id,string"`
	DeviceName   string   `json:"device_name"`
	PlatformType string   `json:"platform_type"` // android, ios, windows
	OSVersion    string   `json:"os_version"`
	UDID         string   `json:"udid"`
	SerialNumber string   `json:"serial_number"`
	IMEI         []string `json:"imei"` // デュアルSIM対応で配列
	Model        string   `json:"model"`
	User         *User    `json:"user,omitempty"`
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
	case "ios":
		return "🍎"
	case "android":
		return "🤖"
	case "windows":
		return "💻"
	default:
		return "📱"
	}
}

// PlatformDisplayName はプラットフォームの表示名を返す
func (d *Device) PlatformDisplayName() string {
	switch d.PlatformType {
	case "ios":
		return "iOS"
	case "android":
		return "Android"
	case "windows":
		return "Windows"
	default:
		return d.PlatformType
	}
}

// IMEIString はIMEI配列をカンマ区切りの文字列で返す
func (d *Device) IMEIString() string {
	if len(d.IMEI) == 0 {
		return "-"
	}
	result := ""
	for i, imei := range d.IMEI {
		if i > 0 {
			result += ", "
		}
		result += imei
	}
	return result
}
