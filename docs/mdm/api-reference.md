# ManageEngine MDM API リファレンス

ManageEngine Mobile Device Manager Plus の REST API に関する情報をまとめています。

## 一次情報（公式ドキュメント）

| ドキュメント | URL |
|-------------|-----|
| API Introduction | https://www.manageengine.com/mobile-device-management/api/introduction/ |
| Devices API | https://www.manageengine.com/mobile-device-management/api/devices/ |
| API ドキュメント トップ | https://www.manageengine.com/mobile-device-management/api/ |

## 認証

### オンプレミス vs クラウド

| 環境 | 認証方式 |
|------|---------|
| オンプレミス | APIキー |
| クラウド | Zoho OAuth 2.0 |

### Zoho OAuth 2.0（クラウド版）

#### エンドポイント

| リージョン | Accounts URL |
|-----------|--------------|
| US | https://accounts.zoho.com |
| EU | https://accounts.zoho.eu |
| IN | https://accounts.zoho.in |
| AU | https://accounts.zoho.com.au |
| CN | https://accounts.zoho.com.cn |

#### トークン取得

**Authorization Code → Access Token + Refresh Token**

```
POST {Accounts_URL}/oauth/v2/token
```

パラメータ：
- `code`: Authorization Code
- `client_id`: Client ID
- `client_secret`: Client Secret
- `grant_type`: `authorization_code`

**Refresh Token → Access Token**

```
POST {Accounts_URL}/oauth/v2/token
```

パラメータ：
- `refresh_token`: Refresh Token
- `client_id`: Client ID
- `client_secret`: Client Secret
- `grant_type`: `refresh_token`

レスポンス：
```json
{
  "access_token": "1000.xxxxxxxx",
  "expires_in": 3600,
  "api_domain": "https://www.zohoapis.com",
  "token_type": "Bearer"
}
```

#### トークン有効期限

| トークン | 有効期限 |
|---------|---------|
| Access Token | 1時間（3600秒） |
| Refresh Token | 無期限 |

#### レート制限

- 10分間に最大10個のAccess Tokenを生成可能
- 1ユーザーあたり最大20個のRefresh Token
- 1 Refresh Tokenあたり最大30個のアクティブなAccess Token

---

## デバイス API

### デバイス一覧取得

```
GET /api/v1/mdm/devices
```

**OAuth Scope**: `MDMOnDemand.MDMInventory.READ`

#### リクエストヘッダー

```
Authorization: Zoho-oauthtoken {access_token}
```

#### クエリパラメータ

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `include_all` | boolean | 全デバイスを表示（削除済み含む） |
| `search` | string | デバイス名で検索 |
| `group_id` | integer | グループIDでフィルタ |
| `exclude_removed` | boolean | 削除・未管理・廃止デバイスを除外 |
| `imei` | string | IMEI番号でフィルタ |
| `owned_by` | integer | 所有形態（1: 企業、2: 個人） |
| `device_type` | string | デバイス種別（smartphone, Tablet等） |
| `serial_number` | string | シリアル番号でフィルタ |
| `email` | string | ユーザーのメールアドレス |
| `platform` | string | プラットフォーム（iOS, Android, Windows） |

#### レスポンス

```json
{
  "devices": [
    {
      "device_id": 123456789,
      "device_name": "iPhone-001",
      "platform_type": "ios",
      "os_version": "17.2",
      "udid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
      "serial_number": "ABCD1234EFGH",
      "imei": "123456789012345",
      "model": "iPhone 15 Pro",
      "user": {
        "user_id": 12345,
        "user_name": "田中太郎",
        "user_email": "tanaka@example.com"
      },
      "summary": {
        "profile_count": "3",
        "app_count": "25",
        "doc_count": "0",
        "group_count": "2"
      }
    }
  ]
}
```

#### レスポンスフィールド

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `device_id` | integer | デバイスの一意識別子 |
| `device_name` | string | デバイス名 |
| `platform_type` | string | プラットフォーム（ios, android, windows） |
| `os_version` | string | OSバージョン |
| `udid` | string | デバイスのUDID |
| `serial_number` | string | シリアル番号 |
| `imei` | string | IMEI番号 |
| `model` | string | デバイスモデル |
| `user` | object | 関連付けられたユーザー情報 |
| `summary` | object | プロファイル・アプリ等のカウント |

### デバイス詳細取得

```
GET /api/v1/mdm/devices/{device_id}
```

**OAuth Scope**: `MDMOnDemand.MDMInventory.READ`

#### パスパラメータ

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `device_id` | integer | デバイスID |

#### クエリパラメータ

| パラメータ | 型 | 説明 |
|-----------|-----|------|
| `include_summary` | boolean | サマリー情報を含める |

### デバイス情報更新

```
PUT /api/v1/mdm/devices/{device_id}
```

**OAuth Scope**: `MDMOnDemand.MDMDeviceMgmt.CREATE`

デバイス名、アセットタグ等を更新できます。

---

## スコープ一覧

| スコープ | 説明 |
|---------|------|
| `MDMOnDemand.MDMInventory.READ` | デバイス情報の読み取り |
| `MDMOnDemand.MDMDeviceMgmt.CREATE` | デバイス情報の更新 |

---

## curlでのテスト例

### デバイス一覧取得

```bash
curl -X GET "https://mdm.manageengine.com/api/v1/mdm/devices" \
  -H "Authorization: Zoho-oauthtoken YOUR_ACCESS_TOKEN"
```

### Androidデバイスのみ取得

```bash
curl -X GET "https://mdm.manageengine.com/api/v1/mdm/devices?platform=Android" \
  -H "Authorization: Zoho-oauthtoken YOUR_ACCESS_TOKEN"
```

### デバイス名で検索

```bash
curl -X GET "https://mdm.manageengine.com/api/v1/mdm/devices?search=iPhone" \
  -H "Authorization: Zoho-oauthtoken YOUR_ACCESS_TOKEN"
```

---

---

## アプリ管理 API

### アプリ一覧取得

```
GET /api/v1/mdm/apps
```

**OAuth Scope**: `MDMOnDemand.MDMDeviceMgmt.READ`

#### レスポンス

```json
{
  "apps": [
    {
      "app_id": 123456789,
      "app_name": "My App",
      "app_category": "Business",
      "app_type": 0,
      "version": "1.0.0",
      "platform_type": 2,
      "description": "アプリの説明",
      "icon": "https://...",
      "added_time": 1701234567890,
      "modified_time": 1701234567890,
      "release_labels": []
    }
  ]
}
```

#### レスポンスフィールド

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `app_id` | integer | アプリの一意識別子 |
| `app_name` | string | アプリ名 |
| `app_category` | string | カテゴリ |
| `app_type` | integer | 0=無料, 1=有料, 2=エンタープライズ |
| `version` | string | バージョン |
| `platform_type` | integer | 1=iOS, 2=Android, 3=Windows |
| `icon` | string | アイコンURL |
| `added_time` | integer | 追加日時（Unix時間ミリ秒） |
| `release_labels` | array | リリースラベル（バージョン管理用） |

### アプリをデバイスに配布

```
POST /api/v1/mdm/apps/{app_id}/labels/{release_label_id}/devices
```

**OAuth Scope**: `MDMOnDemand.MDMDeviceMgmt.CREATE`

#### リクエストボディ

```json
{
  "device_ids": [123456, 789012],
  "silent_install": true,
  "notify_user_via_email": false
}
```

| パラメータ | 型 | 必須 | 説明 |
|-----------|-----|------|------|
| `device_ids` | array | ✓ | デバイスIDの配列 |
| `silent_install` | boolean | - | サイレントインストール（ユーザー操作不要） |
| `notify_user_via_email` | boolean | - | メール通知を送信 |

#### レスポンス

成功時: `HTTP 202 Accepted`

### アプリをグループに配布

```
POST /api/v1/mdm/apps/{app_id}/labels/{release_label_id}/groups
```

**OAuth Scope**: `MDMOnDemand.MDMDeviceMgmt.CREATE`

#### リクエストボディ

```json
{
  "group_ids": [111, 222],
  "silent_install": true,
  "notify_user_via_email": false
}
```

---

## 参考リンク

- [Zoho OAuth 2.0 ドキュメント](https://www.zoho.com/accounts/protocol/oauth.html)
- [Zoho API Console](https://api-console.zoho.com/)
- [ManageEngine MDM 製品ページ](https://www.manageengine.com/mobile-device-management/)
- [ManageEngine MDM Apps API](https://www.manageengine.com/mobile-device-management/api/apps/)
