# TechBrew

複数の技術ブログ・ニュースサイト（RSS）から記事を**並行取得**し、ソース別に整理した HTML ダイジェストを生成するツールです。

**推奨:** `--no-ai` で RSS の原文テキストを表示し、気になった記事は元サイトで読む使い方（API キー・従量課金不要）。

## 機能

- Go の goroutine + semaphore による並行 RSS 取得
- ソース別の記事カード表示（タイトル・日時・RSS 概要・元記事リンク）
- `~/TechDigest/latest.html` への HTML 出力（日付別コピーも保存）
- ローカル HTTP サーバー（`localhost:8080`）でダイジェスト表示
- macOS `launchd` による毎朝 6:00 の自動実行テンプレート
- （任意）OpenAI API による日本語 3 トピック要約

## 必要条件

- Go 1.22+

OpenAI API キーは **任意** です。使わない場合は `--no-ai` を指定してください。

## セットアップ

```bash
cd techbrew
go mod download
```

API 要約を使う場合のみ `.env` を用意します。

```bash
cp .env.example .env
# OPENAI_API_KEY を設定
export $(grep -v '^#' .env | xargs)
```

## 使い方

### 1. ダイジェスト生成（推奨: --no-ai）

```bash
go run ./cmd/curator --config config/sites.yaml --no-ai
```

生成先: `~/TechDigest/latest.html`（`OUTPUT_PATH` で変更可）

実行ログの例:

```
[1/4] 設定を読み込み中...
[2/4] RSS を並行取得中...
      ✓ Qiita: 5 件
      ...
[3/4] AI 要約をスキップ（--no-ai）
[4/4] HTML を書き出し中...
```

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

ビルド後はプロジェクト外からでも実行できます（`config/sites.yaml` を自動探索）。

```bash
~/IdeaProjects/techbrew/bin/curator --config config/sites.yaml --no-ai
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
| `OUTPUT_PATH` | `~/TechDigest/latest.html` | 出力 HTML パス |
| `MAX_CONCURRENCY` | `5` | 同時 RSS 取得数 |
| `OPENAI_API_KEY` | — | （任意）AI 要約を使う場合のみ |
| `OPENAI_MODEL` | `gpt-4o-mini` | （任意）使用モデル |

### CLI フラグ

| フラグ | 説明 |
|--------|------|
| `--config` | RSS 設定ファイル（デフォルト: `config/sites.yaml`） |
| `--no-ai` | AI 要約をスキップし RSS 原文を表示（**推奨**） |

## macOS 定期実行（毎朝 6:00）

1. バイナリをビルド: `go build -o bin/curator ./cmd/curator`
2. `deploy/com.user.techbrew.plist` 内のパスを自分の環境に合わせて編集
3. ログ用ディレクトリ作成: `mkdir -p ~/TechDigest/logs`
4. インストール:

```bash
cp deploy/com.user.techbrew.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.user.techbrew.plist
```

plist の `ProgramArguments` には `--no-ai` を含めることを推奨します。

### スリープ時の注意

Mac が 6:00 にスリープ中の場合、ジョブは起動後に実行されるか、翌日までスキップされることがあります。確実に毎朝取得したい場合は、電源を入れたままにするか、起床後に手動実行してください。

```bash
go run ./cmd/curator --config config/sites.yaml --no-ai
```

## 朝の見方（おすすめ）

1. 前夜に `cmd/serve` を起動（またはログイン時起動に登録）
2. ブラウザで http://localhost:8080 をピン留め
3. 毎朝 6:00 に `curator --no-ai` が自動実行され、タブを開けば最新ダイジェストが表示される

## AI 要約について（任意）

`--no-ai` を付けない場合、OpenAI API で「今日の重要トピック 3 選」を日本語要約します。ただし:

- API キーと従量課金が必要
- 要約だけで済ませず、**元記事を読む**ことを推奨

通常利用では `--no-ai` で十分です。

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
│   ├── summarizer/  # LLM 要約（任意）
│   └── report/      # HTML 生成
├── config/sites.yaml
└── deploy/          # launchd テンプレート
```

## ライセンス

MIT
