# TechBrew

複数の技術ブログ・ニュースサイト（RSS）から記事を**並行取得**し、ソース別に整理した HTML ダイジェストを生成するツールです。

**デフォルトは RSS 原文表示**（API キー不要・従量課金なし）。気になった記事は元サイトで読む使い方を推奨します。

**作者より ☕️** 要約は自分では使っていません。せっかく書いてくれた記事は、ちゃんと読みたい派です。

> Go と並行処理の勉強用に作りました。実装では [Cursor](https://cursor.com/) を使っています。

## 機能

- Go の goroutine + semaphore による並行 RSS 取得
- ソース別の記事カード表示（タイトル・日時・RSS 概要・元記事リンク）
- `~/TechDigest/latest.html` への HTML 出力（日付別コピーも保存）
- ローカル HTTP サーバー（`localhost:8080`）でダイジェスト表示
- macOS `launchd` による毎朝 6:00 の自動実行テンプレート
- （任意・opt-in）`--ai` で OpenAI による日本語要約

## 必要条件

- Go 1.22+

OpenAI API キーは **`--ai` を使う場合のみ** 必要です。

## セットアップ

```bash
cd techbrew
go mod download
```

## 使い方

### 1. ダイジェスト生成（デフォルト）

```bash
go run ./cmd/curator --config config/sites.yaml
```

生成先: `~/TechDigest/latest.html`（`OUTPUT_PATH` で変更可）

実行ログの例:

```
[1/4] 設定を読み込み中...
[2/4] RSS を並行取得中...
      ✓ Qiita: 5 件
      ...
[3/4] RSS 原文表示（デフォルト）
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
~/IdeaProjects/techbrew/bin/curator --config config/sites.yaml
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
| `OPENAI_API_KEY` | — | `--ai` 使用時のみ必須 |
| `OPENAI_MODEL` | `gpt-4o-mini` | `--ai` 使用時のモデル |

### CLI フラグ

| フラグ | 説明 |
|--------|------|
| `--config` | RSS 設定ファイル（デフォルト: `config/sites.yaml`） |
| `--ai` | **opt-in:** OpenAI 要約を有効化（キー・料金・送信内容は利用者責任） |
| `--no-ai` | 非推奨（互換用）。デフォルトが RSS のため効果なし |

## macOS 定期実行（毎朝 6:00）

1. バイナリをビルド: `go build -o bin/curator ./cmd/curator`
2. `deploy/com.user.techbrew.plist` 内のパスを自分の環境に合わせて編集
3. ログ用ディレクトリ作成: `mkdir -p ~/TechDigest/logs`
4. インストール:

```bash
cp deploy/com.user.techbrew.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.user.techbrew.plist
```

デフォルトで RSS のみ実行されます（`--ai` は付けません）。

### スリープ時の注意

Mac が 6:00 にスリープ中の場合、ジョブは起動後に実行されるか、翌日までスキップされることがあります。

```bash
go run ./cmd/curator --config config/sites.yaml
```

## 朝の見方（おすすめ）

1. 前夜に `cmd/serve` を起動（またはログイン時起動に登録）
2. ブラウザで http://localhost:8080 をピン留め
3. 毎朝 6:00 に `curator` が自動実行され、タブを開けば最新ダイジェストが表示される

## AI 要約（`--ai`・任意）

明示的に `--ai` を付けた場合のみ OpenAI API を呼び出します。

```bash
cp .env.example .env
# OPENAI_API_KEY を設定
export $(grep -v '^#' .env | xargs)
go run ./cmd/curator --config config/sites.yaml --ai
```

- 記事テキストが **OpenAI に送信** されます
- API 料金・キー管理は **利用者の責任**
- 要約は参考情報。**元記事を必ず確認** してください

詳細は [DISCLAIMER.md](DISCLAIMER.md) を参照してください。

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
│   ├── summarizer/  # LLM 要約（--ai 時のみ）
│   └── report/      # HTML 生成
├── config/sites.yaml
├── DISCLAIMER.md    # 利用上の注意
└── deploy/          # launchd テンプレート
```

## ライセンス

MIT — 利用上の注意は [DISCLAIMER.md](DISCLAIMER.md) を参照
