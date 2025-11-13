# RSS Discord Notifier

RSSフィードを定期的に監視し、新しい記事をDiscordチャンネルに自動通知するGo製のツールです。GitHub Actionsで動作し、誰でも簡単にCloneして使用できます。

## ✨ 特徴

- 🚀 **シンプルなセットアップ**: 5分以内で稼働開始
- 🔄 **自動監視**: GitHub Actionsで定期実行（cron設定可能）
- 📰 **複数フィード対応**: 複数のRSSフィードを同時に監視
- 🎨 **美しい通知**: Discord Embedsを使用した視認性の高い通知
- 🔒 **セキュア**: Webhook URLをGitHub Secretsで安全に管理
- 🌐 **汎用的**: 他のユーザーがCloneしてすぐに使える設計
- 🧩 **拡張可能**: フィードの追加やカスタマイズが容易

## 📋 目次

- [クイックスタート](#-クイックスタート)
- [必要要件](#-必要要件)
- [セットアップ](#️-セットアップ)
- [設定](#️-設定)
- [使い方](#-使い方)
- [開発](#-開発)
- [アーキテクチャ](#-アーキテクチャ)
- [トラブルシューティング](#-トラブルシューティング)
- [ライセンス](#-ライセンス)

## 🚀 クイックスタート

```bash
# 1. リポジトリをClone
git clone https://github.com/yourusername/rss-discord-notifier.git
cd rss-discord-notifier

# 2. 設定ファイルをコピー
cp configs/feeds.example.yaml configs/feeds.yaml

# 3. feeds.yamlを編集して、監視したいRSSフィードを追加

# 4. Discord Webhook URLを作成（後述）

# 5. GitHub Secretsに設定（後述）

# 6. GitHub Actionsを有効化
# Settings > Actions > General > "Allow all actions and reusable workflows"

# 完了！次回のスケジュール実行を待つか、手動実行
```

## 📦 必要要件

### 実行環境（GitHub Actions）
- GitHub アカウント
- Discord サーバーとチャンネル

### ローカル開発（オプション）
- Go 1.21 以上
- Git

## 🛠️ セットアップ

### 1. Discord Webhook URLの作成

1. Discordサーバーの設定を開く
2. 「連携サービス」→「ウェブフック」を選択
3. 「新しいウェブフック」をクリック
4. 名前とチャンネルを設定
5. 「ウェブフックURLをコピー」

![Discord Webhook Setup](https://docs.discord.com/assets/webhooks_create.png)

### 2. GitHub Secretsの設定

1. GitHubリポジトリの「Settings」タブを開く
2. 「Secrets and variables」→「Actions」を選択
3. 「New repository secret」をクリック
4. 以下のSecretを追加：

| Name | Value | 説明 |
|------|-------|------|
| `DISCORD_WEBHOOK_URL` | `https://discord.com/api/webhooks/...` | 手順1で取得したWebhook URL |

### 3. RSSフィードの設定

`configs/feeds.yaml` を編集して、監視したいRSSフィードを追加します：

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
  
  - name: "あなたのお気に入りのブログ"
    url: "https://example.com/feed"
    category: "Blog"
    enabled: true
```

#### フィールド説明

- `name`: フィードの表示名（Discord通知に表示されます）
- `url`: RSSフィードのURL
- `category`: カテゴリ（`Tech`, `News`, `Blog`, `Other`）。色分けに使用されます
- `enabled`: `true`で有効、`false`で無効

### 4. GitHub Actionsの設定

`.github/workflows/notify.yml`でスケジュールを調整できます：

```yaml
on:
  schedule:
    - cron: '0 * * * *'  # 毎時0分に実行（UTC）
  workflow_dispatch:      # 手動実行も可能
```

#### cronスケジュール例

| cron式 | 実行間隔 |
|--------|----------|
| `0 * * * *` | 1時間ごと |
| `*/30 * * * *` | 30分ごと |
| `0 */2 * * *` | 2時間ごと |
| `0 9,12,15,18 * * *` | 9時、12時、15時、18時（UTC） |

**注意**: GitHub Actionsのcronは数分遅延する場合があります。

## ⚙️ 設定

### 環境変数

| 変数名 | 必須 | デフォルト | 説明 |
|-------|------|-----------|------|
| `DISCORD_WEBHOOK_URL` | ✅ | - | Discord Webhook URL |
| `CONFIG_FILE_PATH` | ❌ | `./configs/feeds.yaml` | 設定ファイルのパス |
| `STATE_FILE_PATH` | ❌ | `./state/state.json` | 状態ファイルのパス |
| `LOG_LEVEL` | ❌ | `INFO` | ログレベル（`DEBUG`, `INFO`, `WARN`, `ERROR`） |
| `LOG_FORMAT` | ❌ | `json` | ログフォーマット（`json`, `text`） |

### カスタマイズ

#### 通知メッセージのカスタマイズ

カテゴリごとの色設定は `internal/discord/message.go` で定義されています：

```go
colors := map[string]int{
    "Tech":  5793522,  // Discord Blurple
    "News":  5763719,  // Green
    "Blog":  16770908, // Yellow
    "Other": 15418782, // Pink
}
```

#### 初回実行時の挙動

初回実行時（状態ファイルがない場合）は、最新5件のみを通知します。過去の全記事が一度に通知されることを防ぎます。

この設定は `internal/state/manager.go` で調整可能です。

## 📖 使い方

### 自動実行（推奨）

GitHub Actionsが設定されたスケジュールで自動実行します。何もする必要はありません！

### 手動実行

GitHub リポジトリの「Actions」タブから手動実行できます：

1. 「Actions」タブを開く
2. 「RSS to Discord Notification」ワークフローを選択
3. 「Run workflow」をクリック

### ローカル実行（開発・テスト用）

```bash
# 依存関係のインストール
go mod download

# 環境変数を設定
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/..."

# 実行
go run cmd/notifier/main.go

# またはビルドして実行
make build
./bin/notifier
```

## 👨‍💻 開発

### プロジェクト構造

```
rss-discord-notifier/
├── cmd/notifier/           # エントリーポイント
├── internal/               # 内部パッケージ
│   ├── config/            # 設定管理
│   ├── feed/              # RSSフィード取得
│   ├── discord/           # Discord通知
│   ├── state/             # 状態管理
│   └── logger/            # ロガー
├── pkg/models/            # データモデル
├── configs/               # 設定ファイル
├── docs/                  # ドキュメント
│   ├── requirements/      # 要件定義・アーキテクチャ
│   └── adr/              # アーキテクチャ決定記録
└── .github/workflows/     # GitHub Actions
```

### ビルド

```bash
# ビルド
make build

# テスト
make test

# Linter実行
make lint

# すべてのチェック
make all
```

### テスト

```bash
# 全テスト実行
go test ./...

# カバレッジ付き
go test -cover ./...

# 詳細出力
go test -v ./...
```

### 依存関係

主要なライブラリ：

- [gofeed](https://github.com/mmcdole/gofeed) - RSSパーサー
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML解析

## 🏗️ アーキテクチャ

詳細なアーキテクチャ設計は以下のドキュメントを参照してください：

- [要件定義書](docs/requirements/requirements.md)
- [アーキテクチャ設計書](docs/requirements/architecture.md)
- [ADR 0001: Go言語の採用](docs/adr/0001-use-go-language.md)
- [ADR 0002: 状態管理方法](docs/adr/0002-state-management.md)
- [ADR 0003: 通知フォーマット](docs/adr/0003-discord-notification-format.md)

### システムフロー

```
GitHub Actions（定期実行）
    ↓
1. 設定ファイル読み込み（feeds.yaml）
    ↓
2. 状態ファイル読み込み（state.json）
    ↓
3. RSSフィード取得（並行処理）
    ↓
4. 新規記事のフィルタリング
    ↓
5. Discord通知送信
    ↓
6. 状態ファイル保存
```

## 🔍 トラブルシューティング

### 通知が届かない

**原因1**: Discord Webhook URLが正しく設定されていない
- GitHub Secretsの`DISCORD_WEBHOOK_URL`を確認
- Webhook URLが有効か確認（Discordの設定から）

**原因2**: RSSフィードのURLが間違っている
- `configs/feeds.yaml`のURLを確認
- ブラウザでURLにアクセスしてXMLが表示されるか確認

**原因3**: GitHub Actionsが実行されていない
- リポジトリの「Actions」タブで実行履歴を確認
- Actions が有効化されているか確認

### 重複通知が発生する

**原因**: 状態ファイルが正しく保存/読み込みされていない
- GitHub Actions のログで Cache/Artifacts の動作を確認
- 手動で状態ファイルを削除して再初期化

### RSSフィードが取得できない

**原因1**: タイムアウト
- `configs/feeds.yaml`の`timeout_seconds`を増やす

**原因2**: フィードのフォーマットが非対応
- ブラウザでフィードを開いて、正しいRSS/Atom形式か確認

### ログの確認方法

GitHub Actions のログを確認：

1. リポジトリの「Actions」タブを開く
2. 該当のワークフロー実行を選択
3. 「Run notifier」ステップのログを確認

ローカル実行時：

```bash
# DEBUGレベルで実行
LOG_LEVEL=DEBUG go run cmd/notifier/main.go
```

## 🤝 コントリビューション

プルリクエストを歓迎します！

1. このリポジトリをFork
2. Feature ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチをプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

### コーディングガイドライン

- Go 標準のコーディングスタイルに従う
- `gofmt`でフォーマット
- ユニットテストを追加
- ADRで重要な設計決定を記録

## 📝 ライセンス

MIT License - 詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 🙏 謝辞

- [gofeed](https://github.com/mmcdole/gofeed) - 優れたRSSパーサーライブラリ
- [Discord](https://discord.com/) - Webhook APIの提供
- [GitHub Actions](https://github.com/features/actions) - 無料のCI/CDプラットフォーム

## 📞 サポート

問題が発生した場合：

1. [Issues](https://github.com/yourusername/rss-discord-notifier/issues)で既存の問題を検索
2. 新しいIssueを作成（再現手順、ログ、環境情報を含めてください）

---

**作成者**: [Your Name](https://github.com/yourusername)  
**プロジェクトリンク**: [https://github.com/yourusername/rss-discord-notifier](https://github.com/yourusername/rss-discord-notifier)
