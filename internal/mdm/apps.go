package mdm

import "time"

// App はMDMに登録されたアプリ情報
type App struct {
	AppID         int64          `json:"app_id"`
	AppName       string         `json:"app_name"`
	AppCategory   string         `json:"app_category"`
	AppType       int            `json:"app_type"` // 0=無料, 1=有料, 2=エンタープライズ
	Version       string         `json:"version"`
	PlatformType  int            `json:"platform_type"` // 1=iOS, 2=Android, 3=Windows
	Description   string         `json:"description"`
	Icon          string         `json:"icon"`
	AddedTime     int64          `json:"added_time"`
	ModifiedTime  int64          `json:"modified_time"`
	ReleaseLabels []ReleaseLabel `json:"release_labels"`
}

// ReleaseLabel はアプリのリリースラベル（バージョン管理用）
type ReleaseLabel struct {
	ReleaseLabelID   int64  `json:"release_label_id"`
	ReleaseLabelName string `json:"release_label_name"`
	Version          string `json:"version"`
}

// AppsResponse はアプリ一覧APIのレスポンス
type AppsResponse struct {
	Apps []App `json:"apps"`
}

// DistributeAppsRequest はアプリ配布リクエスト
type DistributeAppsRequest struct {
	DeviceIDs          []int64 `json:"device_ids,omitempty"`
	GroupIDs           []int64 `json:"group_ids,omitempty"`
	SilentInstall      bool    `json:"silent_install"`
	NotifyUserViaEmail bool    `json:"notify_user_via_email"`
}

// AppTypeString はapp_typeの表示名を返す
func (a *App) AppTypeString() string {
	switch a.AppType {
	case 0:
		return "無料"
	case 1:
		return "有料"
	case 2:
		return "エンタープライズ"
	default:
		return "不明"
	}
}

// PlatformString はplatform_typeの表示名を返す
func (a *App) PlatformString() string {
	switch a.PlatformType {
	case 1:
		return "iOS"
	case 2:
		return "Android"
	case 3:
		return "Windows"
	default:
		return "不明"
	}
}

// PlatformIcon はplatform_typeに応じたアイコンを返す
func (a *App) PlatformIcon() string {
	switch a.PlatformType {
	case 1:
		return "🍎"
	case 2:
		return "🤖"
	case 3:
		return "💻"
	default:
		return "📱"
	}
}

// AddedTimeFormatted は追加日時をフォーマットして返す
func (a *App) AddedTimeFormatted() string {
	if a.AddedTime == 0 {
		return "-"
	}
	return time.UnixMilli(a.AddedTime).Format("2006/01/02 15:04")
}
