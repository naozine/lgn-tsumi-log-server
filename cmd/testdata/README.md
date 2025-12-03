# テストデータ生成ツール

コースの停車地一覧から、Google Maps Directions API を使って実際の道路に沿った `location_logs` を生成するCLIツールです。

## 事前準備

### 1. Google Maps API キーの取得

1. [Google Cloud Console](https://console.cloud.google.com/) にアクセス
2. プロジェクトを作成または選択
3. 「APIとサービス」→「ライブラリ」から **Directions API** を有効化
4. 「APIとサービス」→「認証情報」から API キーを作成

### 2. API キーの設定

以下のいずれかの方法で設定してください。

**方法A: .env ファイルに記載（推奨）**

プロジェクトルートの `.env` ファイルに追加:

```
GOOGLE_MAPS_API_KEY=AIzaSy...your-api-key...
```

**方法B: 環境変数として設定**

```bash
export GOOGLE_MAPS_API_KEY="AIzaSy...your-api-key..."
```

## 使用方法

### 基本的な使い方

```bash
go run cmd/testdata/main.go -project 1 -course "車両1"
```

### すべてのオプション

```bash
go run cmd/testdata/main.go \
  -project 1 \
  -course "車両1" \
  -interval 10 \
  -date "2025-12-02" \
  -db "./app.db"
```

## コマンドライン引数

| 引数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `-project` | ✓ | - | プロジェクトID |
| `-course` | ✓ | - | コース名 |
| `-interval` | - | `10` | ログ生成間隔（秒） |
| `-date` | - | 今日 | 基準日（YYYY-MM-DD形式） |
| `-db` | - | `./app.db` | DBファイルパス |

## 実行例

```
$ go run cmd/testdata/main.go -project 1 -course "車両1"

=== テストデータ生成ツール ===
プロジェクト: 1
コース: 車両1
ログ間隔: 10秒
基準日: 2025-12-02

プロジェクト名: テスト案件

停車地を取得中...
  - 5件の停車地を取得

ルートを取得中...
  [1/4] 出発地点 → A商店 ... OK (12ポイント)
  [2/4] A商店 → B工場 ... OK (25ポイント)
  [3/4] B工場 → Cセンター ... OK (18ポイント)
  [4/4] Cセンター → 帰着地点 ... OK (20ポイント)

完了！
  - ルートセグメント: 4
  - 生成されたログ: 135件
```

## 生成されるデータの仕様

### 移動中のログ

- Google Directions API から取得した実際の道路経路に沿ったポイント
- 速度: 距離と時間から計算（±5%のランダム変動あり）
- 方位角: 進行方向から計算
- GPS精度: 5〜15m のランダム値

### 滞在中のログ

- 停車地の `stay_minutes` の間、`interval` 秒ごとにログ生成
- 位置: 停車地座標（±5m程度のGPSノイズ付加）
- 速度: 0〜3 km/h のランダム値

### 時刻の割り当て

- 各停車地の `arrival_time` を基準に時刻を計算
- 出発時刻 = 前の停車地の `arrival_time` + `stay_minutes`
- 移動中のポイントは距離に応じて線形補間

## 注意事項

- **既存のログは削除されます**: 指定したコースの `location_logs` は全て削除されてから再生成されます
- **API利用料金**: Google Directions API は有料です（$5/1000リクエスト）。停車地の数だけAPIコールが発生します
- **座標が必要**: 停車地に緯度経度が設定されていない場合、そのセグメントはスキップされます

## トラブルシューティング

### API returned status: REQUEST_DENIED

API キーが無効または Directions API が有効化されていません。Google Cloud Console で確認してください。

### API returned status: OVER_QUERY_LIMIT

API の利用上限に達しています。しばらく待つか、課金設定を確認してください。

### no routes found

指定された2点間のルートが見つかりません。座標が正しいか確認してください（海上や道路のない場所など）。
