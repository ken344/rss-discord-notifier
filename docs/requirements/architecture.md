# アーキテクチャ設計書

## 1. システムアーキテクチャ概要

```
┌─────────────────────────────────────────────────────┐
│           GitHub Actions (定期実行)                  │
│  ┌───────────────────────────────────────────────┐  │
│  │         RSS Discord Notifier (Go)             │  │
│  │                                               │  │
│  │  ┌──────────┐  ┌──────────┐  ┌───────────┐  │  │
│  │  │  Config  │  │   Feed   │  │  Discord  │  │  │
│  │  │  Loader  │→│  Fetcher │→│  Notifier │  │  │
│  │  └──────────┘  └──────────┘  └───────────┘  │  │
│  │       ↓              ↓              ↓        │  │
│  │  ┌──────────┐  ┌──────────┐  ┌───────────┐  │  │
│  │  │ feeds.   │  │  State   │  │  Logger   │  │  │
│  │  │ yaml     │  │  Manager │  │           │  │  │
│  │  └──────────┘  └──────────┘  └───────────┘  │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
         ↓                                    ↓
    ┌─────────┐                          ┌──────────┐
    │   RSS   │                          │ Discord  │
    │  Feeds  │                          │ Webhook  │
    └─────────┘                          └──────────┘
```

## 2. ディレクトリ構成

```
rss-discord-notifier/
├── .github/
│   └── workflows/
│       └── notify.yml              # GitHub Actions ワークフロー定義
├── cmd/
│   └── notifier/
│       └── main.go                 # エントリーポイント
├── internal/
│   ├── config/
│   │   ├── config.go              # 設定構造体とローダー
│   │   └── config_test.go         # テスト
│   ├── feed/
│   │   ├── fetcher.go             # RSSフィード取得ロジック
│   │   ├── parser.go              # RSS解析
│   │   └── fetcher_test.go        # テスト
│   ├── discord/
│   │   ├── notifier.go            # Discord通知送信
│   │   ├── message.go             # メッセージフォーマット
│   │   └── notifier_test.go       # テスト
│   ├── state/
│   │   ├── manager.go             # 既読状態管理
│   │   └── manager_test.go        # テスト
│   └── logger/
│       └── logger.go              # ロガー設定
├── pkg/
│   └── models/
│       ├── feed.go                # フィード関連のモデル
│       └── article.go             # 記事のモデル
├── configs/
│   ├── feeds.example.yaml         # フィード設定のサンプル
│   └── feeds.yaml                 # 実際のフィード設定（gitignore対象）
├── state/
│   └── .gitkeep                   # 状態ファイル保存ディレクトリ
├── docs/
│   ├── requirements/
│   │   ├── requirements.md        # 要件定義書
│   │   └── architecture.md        # このファイル
│   └── adr/
│       ├── 0001-use-go.md         # ADR: Go言語の選択
│       ├── 0002-state-management.md # ADR: 状態管理方法
│       └── 0003-notification-format.md # ADR: 通知フォーマット
├── .gitignore
├── .env.example                   # 環境変数のサンプル
├── go.mod                         # Go モジュール定義
├── go.sum                         # 依存関係のチェックサム
├── Makefile                       # ビルドとテストのコマンド
├── README.md                      # プロジェクト説明
└── LICENSE                        # ライセンス

```

### ディレクトリの説明

#### `/cmd/notifier/`
- アプリケーションのエントリーポイント
- `main.go`で各コンポーネントを初期化し、実行フローを制御

#### `/internal/`
- アプリケーション固有のロジック
- 外部パッケージから直接インポートされることを防ぐ
- 各サブパッケージは単一責任の原則に従う

#### `/pkg/models/`
- 共有データモデル
- 将来的に外部パッケージとして公開可能な汎用的な構造

#### `/configs/`
- 設定ファイルの配置場所
- サンプルファイルと実際の設定ファイルを分離

#### `/state/`
- 既読状態を保存するファイルの配置場所
- GitHub Actionsのcacheまたはartifactsで永続化

## 3. コンポーネント設計

### 3.1 Config Loader (`internal/config`)

**責務**: 設定ファイルと環境変数の読み込み

```go
type Config struct {
    Feeds          []FeedConfig
    Discord        DiscordConfig
    Notifications  NotificationConfig
    StateFile      string
}

type FeedConfig struct {
    Name        string
    URL         string
    Category    string
    Enabled     bool
    WebhookURL  string  // オプション: フィード専用のWebhook URL
}

type DiscordConfig struct {
    WebhookURL  string
}

type NotificationConfig struct {
    MaxArticles     int
    TimeoutSeconds  int
    RateLimit       time.Duration
}
```

### 3.2 Feed Fetcher (`internal/feed`)

**責務**: RSSフィードの取得と解析

**主な機能**:
- 複数フィードの並行取得
- タイムアウト処理
- エラーハンドリング（個別フィードのエラーを隔離）
- RSS/Atom両対応

```go
type Fetcher interface {
    FetchAll(ctx context.Context, feeds []FeedConfig) ([]Article, error)
    Fetch(ctx context.Context, feed FeedConfig) ([]Article, error)
}
```

### 3.3 Discord Notifier (`internal/discord`)

**責務**: Discordへの通知送信

**主な機能**:
- Embed形式のメッセージ作成
- レート制限の考慮
- 送信エラーハンドリング

```go
type Notifier interface {
    SendArticle(ctx context.Context, article Article) error
    SendArticles(ctx context.Context, articles []Article) error
}
```

**Discordメッセージフォーマット**:
```json
{
  "embeds": [{
    "title": "記事タイトル",
    "url": "https://example.com/article",
    "description": "記事の要約（最初の200文字）",
    "color": 5814783,
    "thumbnail": {
      "url": "https://example.com/image.jpg"
    },
    "fields": [
      {
        "name": "フィード",
        "value": "フィード名",
        "inline": true
      },
      {
        "name": "公開日時",
        "value": "2025-11-13 10:00:00",
        "inline": true
      }
    ],
    "timestamp": "2025-11-13T10:00:00Z"
  }]
}
```

**注記**: `thumbnail`フィールドはRSSフィードに画像がある場合のみ含まれます。

### 3.4 State Manager (`internal/state`)

**責務**: 既読記事の管理

**主な機能**:
- 最後に通知した記事のGUIDまたはURLをファイルに保存
- 既読チェック
- 状態ファイルの読み書き（JSON形式）

```go
type Manager interface {
    Load() error
    Save() error
    IsNotified(articleID string) bool
    MarkAsNotified(articleID string) error
}

type State struct {
    LastUpdate      time.Time
    NotifiedArticles map[string]NotifiedArticle
}

type NotifiedArticle struct {
    ID          string
    Title       string
    URL         string
    PublishedAt time.Time
    NotifiedAt  time.Time
}
```

### 3.5 Logger (`internal/logger`)

**責務**: 構造化ログの出力

- レベル別ログ（INFO, WARN, ERROR）
- JSON形式での出力（GitHub Actionsとの親和性）

## 4. データフロー

### 4.1 メイン実行フロー

```
1. 設定読み込み
   ├─ feeds.yaml を読み込み
   ├─ 環境変数（DISCORD_WEBHOOK_URL）を読み込み
   └─ バリデーション

2. 状態管理の初期化
   └─ state.json から既読情報を読み込み

3. RSSフィード取得
   ├─ 各フィードを並行で取得
   ├─ 記事のパース
   └─ エラーハンドリング（個別）

4. 新規記事のフィルタリング
   ├─ 既読チェック
   ├─ 公開日時でソート
   └─ 上限数の適用

5. Discord通知
   ├─ メッセージフォーマット
   ├─ 送信（レート制限考慮）
   └─ エラーハンドリング

6. 状態の保存
   ├─ 通知済み記事をマーク
   └─ state.json に保存
```

### 4.2 エラーハンドリング戦略

- **個別フィードのエラー**: ログ出力して処理継続
- **Discord送信エラー**: リトライ（最大3回、指数バックオフ）
- **状態保存エラー**: エラーログを出力して終了（次回実行時に重複通知のリスク）
- **設定エラー**: 即座に終了

## 5. 設定ファイル仕様

### 5.1 feeds.yaml

```yaml
# RSSフィード設定
version: "1.0"

# 通知設定
notification:
  max_articles_per_run: 10        # 1回の実行で通知する最大記事数
  timeout_seconds: 30              # フィード取得のタイムアウト
  rate_limit_ms: 1000              # Discord通知間隔（ミリ秒）

# RSSフィードリスト
feeds:
  - name: "Go Blog"
    url: "https://go.dev/blog/feed.atom"
    category: "Tech"
    enabled: true
  
  - name: "GitHub Blog"
    url: "https://github.blog/feed/"
    category: "Tech"
    enabled: true
  
  - name: "Example Disabled Feed"
    url: "https://example.com/feed"
    category: "Other"
    enabled: false
```

### 5.2 環境変数

```bash
# 必須
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/xxxxx/yyyyy

# オプション
STATE_FILE_PATH=./state/state.json        # デフォルト: ./state/state.json
LOG_LEVEL=INFO                            # デフォルト: INFO
LOG_FORMAT=json                           # デフォルト: json
CONFIG_FILE_PATH=./configs/feeds.yaml     # デフォルト: ./configs/feeds.yaml
```

## 6. GitHub Actions統合

### 6.1 ワークフロー構成

```yaml
name: RSS to Discord Notification

on:
  schedule:
    - cron: '0 * * * *'  # 1時間ごと
  workflow_dispatch:      # 手動実行も可能

jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Cache state
        uses: actions/cache@v3
        with:
          path: state/
          key: rss-state-${{ github.run_id }}
          restore-keys: rss-state-
      
      - name: Run notifier
        env:
          DISCORD_WEBHOOK_URL: ${{ secrets.DISCORD_WEBHOOK_URL }}
        run: |
          go run cmd/notifier/main.go
      
      - name: Upload state artifact
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: rss-state
          path: state/
```

## 7. セキュリティ考慮事項

- ✅ Discord Webhook URLはGitHub Secretsで管理
- ✅ 設定ファイルに機密情報を含めない
- ✅ HTTPSでのRSSフィード取得
- ✅ タイムアウトによるリソース保護
- ✅ 入力バリデーション（URL、設定値）

## 8. テスト戦略

### 8.1 ユニットテスト
- 各パッケージの主要機能
- モックを使用した外部依存の分離

### 8.2 統合テスト
- 実際のRSSフィードを使用したテスト（オプション）
- テスト用のDiscord Webhookを使用

### 8.3 GitHub Actions でのCI
- プルリクエストごとにテスト実行
- linterチェック（golangci-lint）

## 9. パフォーマンス最適化

- **並行処理**: 複数フィードを並行取得（goroutine）
- **タイムアウト**: レスポンスが遅いフィードで停止しない
- **キャッシュ**: GitHub Actions Cacheで状態を永続化
- **レート制限**: Discord APIの制限を遵守

## 10. 監視とログ

- 構造化ログによる実行状況の追跡
- GitHub Actions のログ出力
- エラー発生時の詳細な情報出力
- 通知成功/失敗のカウント

