# REST API仕様

**ベースURL**: `http://localhost:8080` (本番環境では https)

**認証**: 必要。HTTPリクエストヘッダーに `X-Project-Api-Key` を含める必要があります。APIキーはWeb UIのプロジェクト詳細画面から取得できます。

**Content-Type**: `application/json`

---

## デバイス登録 API

**URL**: `POST /api/v1/devices`

モバイルアプリの初回起動時にデバイスを登録するAPI。登録後、管理者がWeb UIでデバイスにコースを割り当てることで、位置情報・写真の送信が可能になります。

### リクエスト仕様

#### ヘッダー

| ヘッダー名            | 値の例            | 必須 | 説明                                     |
|---------------------|-------------------|------|------------------------------------------|
| `X-Project-Api-Key` | `prj_sk_xxxxxxxx...` | ✓    | プロジェクト共通APIキー                      |

#### リクエストボディ（JSON）

```json
{
  "device_id": "ANDROID_abc123def456",
  "device_name": "配送車両1号"
}
```

#### フィールド説明

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `device_id` | string | ✓ | 端末を一意に識別するID（ANDROID_ID、IDFV等） |
| `device_name` | string | - | 端末の表示名（省略時は管理画面で「名前なし」と表示） |

### レスポンス仕様

#### 成功時（HTTP 200）- 新規登録

```json
{
  "success": true,
  "device_id": "ANDROID_abc123def456",
  "course_name": null,
  "message": "Device registered successfully. Please wait for course assignment."
}
```

#### 成功時（HTTP 200）- 既に登録済み（コース割当あり）

```json
{
  "success": true,
  "device_id": "ANDROID_abc123def456",
  "course_name": "車両1",
  "message": "Device already registered"
}
```

#### レスポンスフィールド説明

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `success` | boolean | 処理結果（true: 成功） |
| `device_id` | string | 登録されたデバイスID |
| `course_name` | string/null | 割り当てられたコース名。未割当の場合はnull |
| `message` | string | 結果メッセージ |

### エラー時（HTTP 400/401/500）

```json
{
  "success": false,
  "error": "device_id is required"
}
```

| HTTPステータス | エラーメッセージ | 原因 |
|--------------|----------------|------|
| 400 | `Invalid request format` | JSONフォーマットが不正 |
| 400 | `device_id is required` | device_idが未指定 |
| 401 | `API key is required` | `X-Project-Api-Key` ヘッダーが未指定 |
| 401 | `Invalid API key` | 指定されたAPIキーが無効または存在しない |
| 500 | `Failed to register device` | サーバー内部エラー |

### 備考

- 同じdevice_idで再度呼び出すと、既存のデバイス情報を返します（更新はしません）
- コースの割り当ては管理者がWeb UIで行います
- コース未割当のデバイスで位置情報・写真APIを呼び出すとエラーになります

---

## 位置情報登録 API

**URL**: `POST /api/v1/locations`

### リクエスト仕様

### ヘッダー

| ヘッダー名            | 値の例            | 必須 | 説明                                     |
|---------------------|-------------------|------|------------------------------------------|
| `X-Project-Api-Key` | `prj_sk_xxxxxxxx...` | ✓    | プロジェクト共通APIキー                      |

### リクエストボディ（JSON）

```json
{
  "device_id": "ANDROID_abc123def456",
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
| `device_id` | string | ✓ | 端末を一意に識別するID（事前にデバイス登録APIで登録済みであること） |
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

### エラー時（HTTP 400/401/404/500）

```json
{
  "success": false,
  "error": "Invalid API key"
}
```

| HTTPステータス | エラーメッセージ | 原因 |
|--------------|----------------|------|
| 400 | `Invalid request format` | JSONフォーマットが不正 |
| 400 | `device_id is required` | device_idが未指定 |
| 400 | `locations array cannot be empty` | locations配列が空 |
| 400 | `No valid locations were recorded` | 全ての位置情報が不正 |
| 401 | `API key is required` | `X-Project-Api-Key` ヘッダーが未指定 |
| 401 | `Invalid API key` | 指定されたAPIキーが無効または存在しない |
| 404 | `Device not registered` | device_idが未登録 |
| 400 | `No course assigned to this device` | デバイスにコースが割り当てられていない |
| 500 | `Internal server error` | サーバー内部エラー |

---

## 写真メタデータ登録 API

**URL**: `POST /api/v1/photos`

写真撮影時にメタデータのみを登録するAPI。写真の実データは後でWiFi接続時に同期することを想定しています。

### リクエスト仕様

#### ヘッダー

| ヘッダー名            | 値の例            | 必須 | 説明                                     |
|---------------------|-------------------|------|------------------------------------------|
| `X-Project-Api-Key` | `prj_sk_xxxxxxxx...` | ✓    | プロジェクト共通APIキー                      |

#### リクエストボディ（JSON）

```json
{
  "device_id": "ANDROID_abc123def456",
  "device_photo_id": "IMG_20251202_123456",
  "latitude": 35.681236,
  "longitude": 139.767125,
  "taken_at": "2025-12-02T15:30:00+09:00"
}
```

#### フィールド説明

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `device_id` | string | ✓ | 端末を一意に識別するID（事前にデバイス登録APIで登録済みであること） |
| `device_photo_id` | string | ✓ | 端末に記録されている写真を特定するためのID |
| `latitude` | float | ✓ | 写真を撮った位置の緯度（WGS84） |
| `longitude` | float | ✓ | 写真を撮った位置の経度（WGS84） |
| `taken_at` | string | ✓ | 撮影日時 ISO 8601形式（例: "2025-12-02T15:30:00+09:00"） |

### レスポンス仕様

#### 成功時（HTTP 200）- 該当地点あり

```json
{
  "success": true,
  "photo_id": 1,
  "matched_stop": {
    "id": 5,
    "sequence": "3",
    "stop_name": "○○商店",
    "address": "東京都渋谷区...",
    "latitude": 35.6815,
    "longitude": 139.7675,
    "distance_meters": 45.2
  },
  "message": "Photo registered and matched to stop: ○○商店"
}
```

#### 成功時（HTTP 200）- 該当地点なし

```json
{
  "success": true,
  "photo_id": 2,
  "matched_stop": null,
  "message": "Photo registered but no matching stop found within threshold"
}
```

#### レスポンスフィールド説明

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `success` | boolean | 処理結果（true: 成功） |
| `photo_id` | integer | 登録された写真メタデータのID |
| `matched_stop` | object/null | 該当する停車地の情報。該当なしの場合はnull |
| `message` | string | 結果メッセージ |

#### matched_stop オブジェクト

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `id` | integer | 停車地ID |
| `sequence` | string | 順番 |
| `stop_name` | string | 停車地名 |
| `address` | string | 住所（任意） |
| `latitude` | float | 停車地の緯度 |
| `longitude` | float | 停車地の経度 |
| `distance_meters` | float | 写真撮影位置から停車地までの距離（メートル） |

### エラー時（HTTP 400/401/404/500）

```json
{
  "success": false,
  "error": "device_id is required"
}
```

| HTTPステータス | エラーメッセージ | 原因 |
|--------------|----------------|------|
| 400 | `Invalid request format` | JSONフォーマットが不正 |
| 400 | `device_id is required` | device_idが未指定 |
| 400 | `device_photo_id is required` | device_photo_idが未指定 |
| 400 | `Invalid taken_at format...` | taken_atの形式が不正 |
| 401 | `API key is required` | `X-Project-Api-Key` ヘッダーが未指定 |
| 401 | `Invalid API key` | 指定されたAPIキーが無効または存在しない |
| 404 | `Device not registered` | device_idが未登録 |
| 400 | `No course assigned to this device` | デバイスにコースが割り当てられていない |
| 500 | `Failed to retrieve route stops` | 停車地取得エラー |
| 500 | `Failed to save photo metadata` | 写真メタデータ保存エラー |

### 備考

- 該当地点の判定には、プロジェクト設定の「到着判定範囲（メートル）」を使用します
- 複数の停車地が範囲内にある場合、最も近い停車地がマッチします

---

## 写真アップロード API

**URL**: `POST /api/v1/photos/upload`

事前に登録された写真メタデータに対して、写真の実データをアップロードするAPI。WiFi接続時などに後から同期することを想定しています。

### リクエスト仕様

#### ヘッダー

| ヘッダー名            | 値の例            | 必須 | 説明                                     |
|---------------------|-------------------|------|------------------------------------------|
| `X-Project-Api-Key` | `prj_sk_xxxxxxxx...` | ✓    | プロジェクト共通APIキー                      |
| `Content-Type` | `multipart/form-data` | ✓    | マルチパートフォーム形式                      |

#### リクエストボディ（multipart/form-data）

| フィールド | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `device_photo_id` | string | ✓ | 事前に登録済みのdevice_photo_id |
| `photo` | file | ✓ | 写真ファイル（JPEG/PNG） |

#### curlでの例

```bash
curl -X POST http://localhost:8080/api/v1/photos/upload \
  -H "X-Project-Api-Key: prj_sk_xxxxxxxx..." \
  -F "device_photo_id=IMG_20251202_123456" \
  -F "photo=@/path/to/photo.jpg"
```

### レスポンス仕様

#### 成功時（HTTP 200）

```json
{
  "success": true,
  "photo_id": 1,
  "file_path": "photos/1/車両1/IMG_20251202_123456.jpg",
  "message": "Photo uploaded successfully"
}
```

#### レスポンスフィールド説明

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `success` | boolean | 処理結果（true: 成功） |
| `photo_id` | integer | 写真メタデータのID |
| `file_path` | string | 保存されたファイルの相対パス |
| `message` | string | 結果メッセージ |

### エラー時（HTTP 400/401/404/409/500）

```json
{
  "success": false,
  "error": "Photo already uploaded"
}
```

| HTTPステータス | エラーメッセージ | 原因 |
|--------------|----------------|------|
| 400 | `device_photo_id is required` | device_photo_idが未指定 |
| 400 | `photo file is required` | 写真ファイルが未指定 |
| 400 | `Unsupported file format. Use JPEG or PNG` | 対応していないファイル形式 |
| 401 | `API key is required` | `X-Project-Api-Key` ヘッダーが未指定 |
| 401 | `Invalid API key` | 指定されたAPIキーが無効または存在しない |
| 404 | `Photo metadata not found...` | 事前にメタデータが登録されていない |
| 409 | `Photo already uploaded` | 既にアップロード済み |
| 500 | `Failed to create storage directory` | ストレージディレクトリ作成失敗 |
| 500 | `Failed to save file` | ファイル保存失敗 |

### 備考

- 写真のアップロード前に、必ず `POST /api/v1/photos` でメタデータを登録してください
- 対応形式: JPEG (.jpg, .jpeg), PNG (.png)
- ファイルサイズ制限: なし（サーバー設定に依存）
- 同じdevice_photo_idで再アップロードするとエラー（409 Conflict）になります
- ファイルは `data/photos/{project_id}/{course_name}/{device_photo_id}.{ext}` に保存されます

---

## API利用フロー

モバイルアプリからAPIを利用する一般的なフローは以下の通りです：

```
1. アプリ初回起動
   └─> POST /api/v1/devices （デバイス登録）
       └─> 管理者がWeb UIでコースを割り当てるまで待機

2. コース割当後、位置情報の送信開始
   └─> POST /api/v1/locations （定期的に送信）

3. 荷物積込時に写真撮影
   └─> POST /api/v1/photos （メタデータ登録）
   └─> POST /api/v1/photos/upload （WiFi接続時に写真アップロード）
```
