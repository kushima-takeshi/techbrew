# TechBrew

複数の技術ブログ・ニュースサイト（RSS）から記事を**並行取得**し、LLM で「今日の重要トピック 3 選」を日本語要約して HTML ダイジェストを生成するツールです。

## 機能

- Go の goroutine + semaphore による並行 RSS 取得
- OpenAI API による日本語 3 トピック要約（Summarizer インターフェースで差し替え可能）
- `~/TechDigest/latest.html` への HTML 出力（日付別コピーも保存）
- ローカル HTTP サーバー（`localhost:8080`）でダイジェスト表示
- macOS `launchd` による毎朝 6:00 の自動実行テンプレート

## 必要条件

- Go 1.22+
- OpenAI API キー

## セットアップ

```bash
cd techbrew
go mod download

cp .env.example .env
# .env を編集して OPENAI_API_KEY を設定
export $(grep -v '^#' .env | xargs)
```

## 使い方

### 1. ダイジェスト生成（手動）

```bash
go run ./cmd/curator --config config/sites.yaml
```

生成先: `~/TechDigest/latest.html`（`OUTPUT_PATH` で変更可）

### 2. ローカルサーバーで表示

```bash
go run ./cmd/serve --port 8080
```

ブラウザで http://localhost:8080 を開く。朝のルーティン用にピン留めタブへの追加を推奨。

### 3. バイナリビルド

```bash
mkdir -p bin
go build -o bin/curator ./cmd/curator
go build -o bin/serve ./cmd/serve
```

## 設定

### RSS ソース（`config/sites.yaml`）

```yaml
sources:
  - name: Qiita
    url: https://qiita.com/popular-items/feed
    max_items: 5
```

ソースの追加・URL 変更はこのファイルを編集してください。

### 環境変数

| 変数 | デフォルト | 説明 |
|------|-----------|------|
| `OPENAI_API_KEY` | — | OpenAI API キー（必須） |
| `OPENAI_MODEL` | `gpt-4o-mini` | 使用モデル |
| `MAX_CONCURRENCY` | `5` | 同時 RSS 取得数 |
| `OUTPUT_PATH` | `~/TechDigest/latest.html` | 出力 HTML パス |

## macOS 定期実行（毎朝 6:00）

1. バイナリをビルド: `go build -o bin/curator ./cmd/curator`
2. `deploy/com.user.techbrew.plist` 内のパスを自分の環境に合わせて編集
3. ログ用ディレクトリ作成: `mkdir -p ~/TechDigest/logs`
4. インストール:

```bash
cp deploy/com.user.techbrew.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.user.techbrew.plist
```

### スリープ時の注意

Mac が 6:00 にスリープ中の場合、ジョブは起動後に実行されるか、翌日までスキップされることがあります。確実に毎朝取得したい場合は、電源を入れたままにするか、起床後に手動実行してください。

```bash
go run ./cmd/curator --config config/sites.yaml
```

## 朝の見方（おすすめ）

1. 前夜に `cmd/serve` を起動（またはログイン時起動に登録）
2. ブラウザで http://localhost:8080 をピン留め
3. 毎朝 6:00 に `curator` が自動実行され、タブを開けば最新ダイジェストが表示される

## テスト

```bash
go test ./...
```

## プロジェクト構成

```
techbrew/
├── cmd/curator/     # メイン CLI
├── cmd/serve/       # ローカル HTTP サーバー
├── internal/
│   ├── config/      # YAML・環境変数
│   ├── fetcher/     # 並行 RSS 取得
│   ├── model/       # データ型
│   ├── summarizer/  # LLM 要約
│   └── report/      # HTML 生成
├── config/sites.yaml
└── deploy/          # launchd テンプレート
```

## ライセンス

MIT
