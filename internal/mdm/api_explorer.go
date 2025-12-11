package mdm

// APIEndpoint はAPI Explorerで使用するエンドポイント定義
type APIEndpoint struct {
	ID             string     `json:"id"`
	Category       string     `json:"category"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Method         string     `json:"method"`
	Path           string     `json:"path"`            // パステンプレート（例: /devices/{device_id}/locations）
	Params         []APIParam `json:"params"`          // パスパラメータやクエリパラメータ
	RequiresDevice bool       `json:"requires_device"` // デバイス選択が必要か
}

// APIParam はAPIパラメータ定義
type APIParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // path, query
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Default     string `json:"default,omitempty"`
}

// GetAPIEndpoints は利用可能なAPIエンドポイント一覧を返す
func GetAPIEndpoints() []APIEndpoint {
	return []APIEndpoint{
		// === デバイス ===
		{
			ID:             "devices_list",
			Category:       "デバイス",
			Name:           "デバイス一覧",
			Description:    "MDM管理下の全デバイスを取得",
			Method:         "GET",
			Path:           "/devices",
			RequiresDevice: false,
		},
		{
			ID:             "device_detail",
			Category:       "デバイス",
			Name:           "デバイス詳細",
			Description:    "指定デバイスの詳細情報を取得",
			Method:         "GET",
			Path:           "/devices/{device_id}",
			RequiresDevice: true,
			Params: []APIParam{
				{Name: "summary", Type: "query", Required: false, Description: "概要情報を含める", Default: "true"},
			},
		},
		{
			ID:             "device_location",
			Category:       "デバイス",
			Name:           "位置情報",
			Description:    "デバイスの現在位置（緯度・経度）を取得",
			Method:         "GET",
			Path:           "/devices/{device_id}/locations",
			RequiresDevice: true,
		},
		{
			ID:             "device_location_address",
			Category:       "デバイス",
			Name:           "位置情報（住所付き）",
			Description:    "デバイスの現在位置と住所を取得",
			Method:         "GET",
			Path:           "/devices/{device_id}/locations_with_address",
			RequiresDevice: true,
		},
		{
			ID:             "device_apps",
			Category:       "デバイス",
			Name:           "インストール済みアプリ",
			Description:    "デバイスにインストールされているアプリ一覧",
			Method:         "GET",
			Path:           "/devices/{device_id}/apps",
			RequiresDevice: true,
		},
		{
			ID:             "device_profiles",
			Category:       "デバイス",
			Name:           "適用プロファイル",
			Description:    "デバイスに適用されているプロファイル一覧",
			Method:         "GET",
			Path:           "/devices/{device_id}/profiles",
			RequiresDevice: true,
		},
		{
			ID:             "device_groups",
			Category:       "デバイス",
			Name:           "所属グループ",
			Description:    "デバイスが所属するグループ一覧",
			Method:         "GET",
			Path:           "/devices/{device_id}/groups",
			RequiresDevice: true,
		},
		{
			ID:             "device_certificates",
			Category:       "デバイス",
			Name:           "証明書",
			Description:    "デバイスにインストールされている証明書一覧",
			Method:         "GET",
			Path:           "/devices/{device_id}/certificates",
			RequiresDevice: true,
		},
		// === アプリ ===
		{
			ID:             "apps_list",
			Category:       "アプリ",
			Name:           "アプリ一覧",
			Description:    "MDMに登録されているアプリ一覧",
			Method:         "GET",
			Path:           "/apps",
			RequiresDevice: false,
		},
		{
			ID:             "app_detail",
			Category:       "アプリ",
			Name:           "アプリ詳細",
			Description:    "指定アプリの詳細情報",
			Method:         "GET",
			Path:           "/apps/{app_id}",
			RequiresDevice: false,
			Params: []APIParam{
				{Name: "app_id", Type: "path", Required: true, Description: "アプリID"},
			},
		},
		// === グループ ===
		{
			ID:             "groups_list",
			Category:       "グループ",
			Name:           "グループ一覧",
			Description:    "デバイスグループの一覧",
			Method:         "GET",
			Path:           "/groups",
			RequiresDevice: false,
		},
		{
			ID:             "group_devices",
			Category:       "グループ",
			Name:           "グループ内デバイス",
			Description:    "指定グループに所属するデバイス一覧",
			Method:         "GET",
			Path:           "/groups/{group_id}/devices",
			RequiresDevice: false,
			Params: []APIParam{
				{Name: "group_id", Type: "path", Required: true, Description: "グループID"},
			},
		},
		// === ユーザー ===
		{
			ID:             "users_list",
			Category:       "ユーザー",
			Name:           "ユーザー一覧",
			Description:    "MDMに登録されているユーザー一覧",
			Method:         "GET",
			Path:           "/users",
			RequiresDevice: false,
		},
		{
			ID:             "user_detail",
			Category:       "ユーザー",
			Name:           "ユーザー詳細",
			Description:    "指定ユーザーの詳細情報",
			Method:         "GET",
			Path:           "/users/{user_id}",
			RequiresDevice: false,
			Params: []APIParam{
				{Name: "user_id", Type: "path", Required: true, Description: "ユーザーID"},
			},
		},
		// === プロファイル ===
		{
			ID:             "profiles_list",
			Category:       "プロファイル",
			Name:           "プロファイル一覧",
			Description:    "MDMに登録されているプロファイル一覧",
			Method:         "GET",
			Path:           "/profiles",
			RequiresDevice: false,
		},
	}
}

// GetAPIEndpointByID はIDからエンドポイント定義を取得
func GetAPIEndpointByID(id string) *APIEndpoint {
	for _, ep := range GetAPIEndpoints() {
		if ep.ID == id {
			return &ep
		}
	}
	return nil
}

// GetAPICategories はカテゴリ一覧を取得（順序保持）
func GetAPICategories() []string {
	return []string{"デバイス", "アプリ", "グループ", "ユーザー", "プロファイル"}
}

// GetEndpointsByCategory はカテゴリ別にエンドポイントを取得
func GetEndpointsByCategory() map[string][]APIEndpoint {
	result := make(map[string][]APIEndpoint)
	for _, ep := range GetAPIEndpoints() {
		result[ep.Category] = append(result[ep.Category], ep)
	}
	return result
}
