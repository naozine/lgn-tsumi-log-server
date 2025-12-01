**URL**: `POST /api/v1/locations`

**ベースURL**: `http://localhost:8080` (本番環境では https)

**認証**: 不要（現バージョン。将来的に端末認証を追加予定）

**Content-Type**: `application/json`

## リクエスト仕様

### リクエストボディ（JSON）

```json
{
  "project_id": 1,
  "course_name": "車両1のコース",
  "locations": [
    {
      "latitude": 35.681236,
      "longitude": 139.767125,
      "timestamp": "2025-11-26T10:30:00Z",
      "accuracy": 10.5,
      "speed": 30.5,
      "bearing": 180.0,
      "battery_level": 85
    }
  ]
}
```

### フィールド説明

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `project_id` | integer | ✓ | プロジェクトID |
| `course_name` | string | ✓ | コース名（例: "車両1のコース"） |
| `locations` | array | ✓ | 位置情報の配列（1件以上） |

### locations配列の各要素

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `latitude` | float | ✓ | 緯度（WGS84） |
| `longitude` | float | ✓ | 経度（WGS84） |
| `timestamp` | string | ✓ | ISO 8601形式（例: "2025-11-26T10:30:00Z"）<br>UTC推奨 |
| `accuracy` | float | - | 位置精度（メートル単位） |
| `speed` | float | - | 速度（km/h） |
| `bearing` | float | - | 方位角（0-360度、北が0度） |
| `battery_level` | integer | - | バッテリー残量（0-100%） |

## レスポンス仕様

### 成功時（HTTP 200）

```json
{
  "success": true,
  "recorded": 2,
  "message": "2 locations recorded"
}
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `success` | boolean | 処理結果（true: 成功） |
| `recorded` | integer | 記録された位置情報の件数 |
| `message` | string | 成功メッセージ |

### エラー時（HTTP 400/404/500）

```json
{
  "success": false,
  "error": "Invalid project_id"
}
```

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `success` | boolean | 処理結果（false: 失敗） |
| `error` | string | エラーメッセージ |

### エラーメッセージ一覧

| HTTPステータス | エラーメッセージ | 原因 |
|--------------|----------------|------|
| 400 | `Invalid request format` | JSONフォーマットが不正 |
| 400 | `project_id is required` | project_idが未指定 |
| 400 | `course_name is required` | course_nameが未指定 |
| 400 | `locations array cannot be empty` | locations配列が空 |
| 400 | `No valid locations were recorded` | 全ての位置情報が不正 |
| 404 | `Invalid project_id` | 指定されたプロジェクトが存在しない |
| 500 | `Internal server error` | サーバー内部エラー |