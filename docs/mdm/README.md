# MDM管理機能

ManageEngine Mobile Device Manager Plus（MDM）と連携し、MDM管理下のデバイス情報をこのアプリケーションから確認できる機能です。

## 機能概要

| 機能 | 説明 | URL |
|------|------|-----|
| MDM管理トップ | MDM機能のトップページ | `/mdm` |
| デバイス一覧 | MDM管理下のデバイス一覧表示 | `/mdm/devices` |

### アクセス権限

- **admin** ロールのみアクセス可能
- サイドメニューに「MDM管理」リンクが表示されます

## セットアップ

### 1. Zoho API Consoleでクライアント登録

1. [Zoho API Console](https://api-console.zoho.com/) にアクセス
2. 「Add Client」をクリック
3. 「Self Client」を選択
4. クライアント名を入力して作成

作成後、以下の情報を控えておきます：
- **Client ID**
- **Client Secret**

### 2. Refresh Tokenの取得

1. Zoho API Consoleで作成したSelf Clientを開く
2. 「Generate Code」タブを選択
3. **Scope** に以下を入力：
   ```
   MDMOnDemand.MDMInventory.READ
   ```
4. 「Create」をクリックしてAuthorization Codeを取得
5. 以下のcurlコマンドでRefresh Tokenを取得：

```bash
curl -X POST "https://accounts.zoho.com/oauth/v2/token" \
  -d "code=取得したAuthorization Code" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "grant_type=authorization_code"
```

レスポンス例：
```json
{
  "access_token": "1000.xxxxxxxx",
  "refresh_token": "1000.yyyyyyyy",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

`refresh_token` の値を控えておきます。

### 3. 環境変数の設定

`.env` ファイルに以下を追加：

```bash
# Zoho OAuth設定
ZOHO_CLIENT_ID=1000.XXXXXXXXXX
ZOHO_CLIENT_SECRET=xxxxxxxxxxxxxxxx
ZOHO_REFRESH_TOKEN=1000.xxxxxxxxxxxxxxxx

# Zohoのデータセンター（省略時はUS）
# 自分のデータセンターの確認方法は下記参照
# US: https://accounts.zoho.com
# EU: https://accounts.zoho.eu
# IN: https://accounts.zoho.in
# AU: https://accounts.zoho.com.au
# CN: https://accounts.zoho.com.cn
# JP: https://accounts.zoho.jp
ZOHO_ACCOUNTS_URL=https://accounts.zoho.com

# MDM API設定
MDM_API_BASE_URL=https://mdm.manageengine.com

# キャッシュ設定（秒、省略時は300秒=5分）
MDM_CACHE_TTL_SECONDS=300
```

#### 自分のZohoデータセンターの確認方法

以下のいずれかの方法で確認できます：

1. **URLで確認**: Zohoにログイン後、ブラウザのURLを確認
   - `accounts.zoho.com` → US
   - `accounts.zoho.eu` → EU
   - `accounts.zoho.in` → IN
   - `accounts.zoho.jp` → JP
   - `accounts.zoho.com.au` → AU

2. **公式ツール**: [Know Your Datacenter](https://www.zoho.com/know-your-datacenter.html) で国ごとのデータセンターを確認

3. **アカウント設定**: Zohoにログイン → 右上のアバター → 「My Account」でデータセンター情報を確認

### 4. 動作確認

1. サーバーを起動
2. adminユーザーでログイン
3. サイドメニューから「MDM管理」を選択
4. デバイス一覧が表示されれば成功

## 認証の仕組み

```
┌─────────────────────────────────────────────────────────────┐
│ 初回セットアップ（手動、1回のみ）                            │
│ Zoho API Console → Refresh Token取得 → .envに保存           │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ アプリ動作時（自動）                                         │
│ Refresh Token → Access Token取得（1時間有効）                │
│ → MDM API呼び出し → 期限切れ時は自動リフレッシュ             │
└─────────────────────────────────────────────────────────────┘
```

- **Refresh Token**: 無期限（.envに保存）
- **Access Token**: 1時間有効（アプリが自動管理）

ユーザーが認証操作を行う必要はありません。

## キャッシュ

MDM APIへのリクエストを減らすため、デバイス一覧はキャッシュされます。

- デフォルトTTL: 5分（300秒）
- `MDM_CACHE_TTL_SECONDS` で変更可能
- キャッシュはインメモリ（サーバー再起動でクリア）

## トラブルシューティング

### 「MDMが設定されていません」と表示される

必要な環境変数がすべて設定されているか確認してください：
- `ZOHO_CLIENT_ID`
- `ZOHO_CLIENT_SECRET`
- `ZOHO_REFRESH_TOKEN`
- `MDM_API_BASE_URL`

### 「oauth error: invalid_code」エラー

Authorization Codeの有効期限が切れています（数分で期限切れ）。
Zoho API Consoleで新しいコードを生成し、すぐにRefresh Tokenを取得してください。

### 「oauth error: invalid_client」エラー

Client IDまたはClient Secretが間違っています。
Zoho API Consoleで正しい値を確認してください。

### デバイス一覧が表示されない

1. MDMにデバイスが登録されているか確認
2. 使用しているZohoアカウントにMDMへのアクセス権限があるか確認
3. スコープ `MDMOnDemand.MDMInventory.READ` が正しく設定されているか確認
