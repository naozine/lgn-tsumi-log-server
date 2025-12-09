# =============================================================================
# Dockerfile - ビルド済みバイナリをコピー
# =============================================================================
# VPS 上で Go ビルド後、バイナリと静的ファイルをコピーするだけの軽量イメージ
# SQLite は modernc.org/sqlite (Pure Go) を使用
# =============================================================================

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# ビルド済みバイナリをコピー（docker-deploy で事前にビルド）
COPY server /app/server

# 静的ファイルをコピー
COPY web/static /app/web/static

# データディレクトリ (SQLite DB用) - ボリュームマウント先
VOLUME ["/app/data"]

# 環境変数のデフォルト値
ENV PORT=8080

# ポートを公開
EXPOSE 8080

# 実行
ENTRYPOINT ["/app/server"]
