package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/database"
	"github.com/naozine/project_crud_with_auth_tmpl/internal/testdata"
	_ "modernc.org/sqlite"
)

func main() {
	// .env ファイルを読み込み
	godotenv.Load()

	// コマンドライン引数
	projectID := flag.Int64("project", 0, "プロジェクトID（必須）")
	courseName := flag.String("course", "", "コース名（必須）")
	interval := flag.Int("interval", 10, "ログ生成間隔（秒）")
	dbPath := flag.String("db", "./app.db", "DBファイルパス")
	dateStr := flag.String("date", "", "基準日（YYYY-MM-DD形式、省略時は今日）")
	flag.Parse()

	// バリデーション
	if *projectID == 0 {
		fmt.Fprintln(os.Stderr, "エラー: -project は必須です")
		flag.Usage()
		os.Exit(1)
	}
	if *courseName == "" {
		fmt.Fprintln(os.Stderr, "エラー: -course は必須です")
		flag.Usage()
		os.Exit(1)
	}

	// Google Maps API キー
	apiKey := os.Getenv("GOOGLE_MAPS_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "エラー: 環境変数 GOOGLE_MAPS_API_KEY が設定されていません")
		os.Exit(1)
	}

	// 基準日
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	baseDate := time.Now().In(jst)
	if *dateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *dateStr, jst)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: 日付形式が不正です: %s\n", *dateStr)
			os.Exit(1)
		}
		baseDate = parsed
	}

	fmt.Println("=== テストデータ生成ツール ===")
	fmt.Printf("プロジェクト: %d\n", *projectID)
	fmt.Printf("コース: %s\n", *courseName)
	fmt.Printf("ログ間隔: %d秒\n", *interval)
	fmt.Printf("基準日: %s\n", baseDate.Format("2006-01-02"))
	fmt.Println()

	// DB接続
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: DB接続に失敗: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	queries := database.New(db)
	ctx := context.Background()

	// プロジェクトの存在確認
	project, err := queries.GetProject(ctx, *projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: プロジェクトが見つかりません: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("プロジェクト名: %s\n", project.Name)
	fmt.Println()

	// 停車地を取得
	fmt.Println("停車地を取得中...")
	stops, err := queries.ListRouteStopsByCourse(ctx, database.ListRouteStopsByCourseParams{
		ProjectID:  *projectID,
		CourseName: *courseName,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 停車地の取得に失敗: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  - %d件の停車地を取得\n", len(stops))
	fmt.Println()

	if len(stops) < 2 {
		fmt.Fprintln(os.Stderr, "エラー: 停車地が2件以上必要です")
		os.Exit(1)
	}

	// Generator を作成
	generator := testdata.NewGenerator(queries, apiKey, *interval)

	// ルートを取得中...
	fmt.Println("ルートを取得中...")
	result, err := generator.Generate(ctx, *projectID, *courseName, baseDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: 生成に失敗: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("完了！\n")
	fmt.Printf("  - ルートセグメント: %d\n", result.RouteSegments)
	fmt.Printf("  - 生成されたログ: %d件\n", result.GeneratedLogs)
}
