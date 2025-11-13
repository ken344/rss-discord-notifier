# RSS Discord Notifier - 開発記録

## プロジェクト概要

GitHub ActionsとGo言語を使用して、RSSフィードを定期的に監視し、新着記事をDiscordに通知するツール。

## 完了した作業（2025-11-13）

### ✅ 要件定義とアーキテクチャ設計

1. **要件定義書の作成** (`docs/requirements/requirements.md`)
   - 機能要件（FR-001〜FR-013）
   - 非機能要件（NFR-001〜NFR-010）
   - 技術スタック（Go、GitHub Actions、主要ライブラリ）
   - ユースケースと成功基準

2. **アーキテクチャ設計書の作成** (`docs/requirements/architecture.md`)
   - システムアーキテクチャ図
   - ディレクトリ構成の設計
   - コンポーネント設計（Config、Feed、Discord、State、Logger）
   - データフローとエラーハンドリング戦略
   - 設定ファイル仕様（feeds.yaml、環境変数）
   - GitHub Actions統合設計

3. **ADR（アーキテクチャ決定記録）の作成**
   - ADR-0001: Go言語の採用理由
   - ADR-0002: 状態管理方法（GitHub Actions Cache + Artifacts）
   - ADR-0003: Discord通知フォーマット（Embeds）
   - ADR README: ADRの概要と管理方針

### ✅ プロジェクト構造の構築

4. **GitHub Actionsワークフローの作成** (`.github/workflows/notify.yml`)
   - Cronスケジュール設定（毎時実行）
   - 手動実行トリガー（workflow_dispatch）
   - 状態管理（Cache + Artifacts のハイブリッド方式）
   - エラー通知機能
   - 実行サマリー表示

5. **設定ファイルとテンプレート**
   - `configs/feeds.example.yaml`: RSSフィード設定のサンプル
   - `.env.example`: 環境変数のテンプレート（未作成：.gitignoreによりブロック）
   - `Makefile`: ビルド、テスト、開発用タスク定義

6. **ドキュメントの整備**
   - `README.md`: プロジェクト説明、セットアップ手順、使い方
   - `docs/adr/README.md`: ADR管理ドキュメント
   - `.gitignore`: 更新（機密情報、ビルド成果物の除外）

7. **ディレクトリ構造の準備**
   - `state/.gitkeep`: 状態ファイル用ディレクトリ

### ✅ 機能拡張（2025-11-13 追加実装）

8. **サムネイル画像表示機能の実装**
   - RSS記事に画像がある場合、Discord通知にサムネイル表示
   - `pkg/models/article.go`: `ImageURL`フィールド追加
   - `internal/feed/fetcher.go`: `extractImageURL()`メソッド実装
     - `<image>`タグからの画像URL抽出
     - `<enclosure type="image/*">`からの画像URL抽出
   - `internal/discord/notifier.go`: Discord EmbedのThumbnail設定
   - 画像がない記事でも正常動作（後方互換性維持）

9. **フィードごとのチャンネル振り分け機能**
   - 設定ファイルの`webhook_url`フィールドでフィード専用チャンネル指定可能
   - 環境変数の展開に対応（`${DISCORD_WEBHOOK_URL_TECH}`形式）
   - デフォルトWebhook URLへのフォールバック機能

10. **ドキュメントの更新**
    - `README.md`: サムネイル画像とチャンネル振り分け機能を特徴に追加
    - `docs/requirements/requirements.md`: FR-007-1（サムネイル）、FR-009（チャンネル振り分け）追加
    - `docs/requirements/architecture.md`: FeedConfigにWebhookURL追加、メッセージフォーマットにthumbnail追加
    - `docs/adr/0003-discord-notification-format.md`: サムネイル画像を「実装済み」に更新
    - `AGENTS.md`: 開発ワークフローとコミット前チェックリストを追加

## 設計のハイライト

### 採用技術
- **言語**: Go 1.21+
- **CI/CD**: GitHub Actions（Cron）
- **通知先**: Discord Webhook
- **状態管理**: GitHub Actions Cache + Artifacts

### 主要な設計判断

1. **Go言語の選択**
   - シングルバイナリで配布が容易
   - 並行処理（goroutine）が強力
   - 標準ライブラリが充実
   - GitHub Actionsとの親和性

2. **状態管理のハイブリッド方式**
   - メイン: GitHub Actions Cache（高速）
   - バックアップ: Artifacts（信頼性）
   - リポジトリを汚さない（自動コミット不使用）

3. **Discord Embedsの採用**
   - 視認性の高い通知
   - カテゴリごとの色分け
   - 構造化された情報表示

### ディレクトリ構成

```
rss-discord-notifier/
├── .github/workflows/    # GitHub Actions
├── cmd/notifier/        # エントリーポイント
├── internal/            # 内部パッケージ
│   ├── config/         # 設定管理
│   ├── feed/           # RSSフィード取得
│   ├── discord/        # Discord通知
│   ├── state/          # 状態管理
│   └── logger/         # ロガー
├── pkg/models/         # データモデル
├── configs/            # 設定ファイル
├── state/              # 状態ファイル
└── docs/               # ドキュメント
```

## 次のステップ（実装フェーズ）

### 1. Goモジュールの初期化
- [ ] `go mod init`
- [ ] 依存関係の追加（gofeed、yaml.v3）

### 2. データモデルの実装
- [ ] `pkg/models/feed.go`
- [ ] `pkg/models/article.go`

### 3. コアパッケージの実装
- [ ] `internal/config/`: 設定読み込み
- [ ] `internal/feed/`: RSSフィード取得
- [ ] `internal/discord/`: Discord通知
- [ ] `internal/state/`: 状態管理
- [ ] `internal/logger/`: ロガー

### 4. メインプログラムの実装
- [ ] `cmd/notifier/main.go`: エントリーポイント

### 5. テストの作成
- [ ] 各パッケージのユニットテスト
- [ ] 統合テスト

### 6. ドキュメントの最終確認
- [ ] README の更新（必要に応じて）
- [ ] セットアップガイドの検証

## プロジェクトの特徴

✨ **汎用性**: 誰でもCloneして使える  
🔒 **セキュア**: 機密情報はGitHub Secretsで管理  
🚀 **シンプル**: 5分でセットアップ完了  
🎨 **美しい通知**: Discord Embedsで視認性向上  
🔄 **自動化**: GitHub Actionsで完全自動  
📝 **ドキュメント充実**: 要件定義からADRまで完備

## メモ

- 初回実行時は最新5件のみ通知（過去の全記事を避ける）
- レート制限対策: 通知間隔1秒
- タイムアウト: フィード取得30秒
- エラーハンドリング: 個別フィードのエラーで全体を止めない設計

---

## 🔄 開発ワークフロー

### コミット前のチェックリスト

**必須: 以下をすべて実行してからコミット・プッシュすること！**

#### 1. コードフォーマット

```bash
# すべてのGoファイルをフォーマット
gofmt -s -w .

# または Makefileを使用
make fmt

# フォーマット確認（何も出力されなければOK）
gofmt -s -l .
```

#### 2. Linter実行

```bash
# golangci-lintを実行
golangci-lint run ./...

# または Makefileを使用
make lint
```

#### 3. 静的解析

```bash
# go vetを実行
go vet ./...

# または Makefileを使用
make vet
```

#### 4. テスト実行

```bash
# すべてのテストを実行
go test ./...

# または Makefileを使用
make test

# カバレッジ付きで実行
make test-coverage
```

#### 5. ビルド確認

```bash
# ビルドが成功するか確認
go build -o bin/notifier cmd/notifier/main.go

# または Makefileを使用
make build
```

#### 6. 依存関係の整理

```bash
# go.modを整理
go mod tidy

# または Makefileを使用
make tidy
```

### 推奨: すべてをまとめて実行

```bash
# すべてのチェックとビルドを一度に実行
make all
```

これで以下が実行されます：
1. ✅ clean（古いビルド成果物を削除）
2. ✅ lint（Linter実行）
3. ✅ test（テスト実行）
4. ✅ build（ビルド）

### Git操作の推奨フロー

```bash
# 1. 変更の確認
git status

# 2. すべてのチェックを実行（重要！）
make all

# 3. フォーマットを確認
gofmt -s -l .

# 4. すべて成功したら、変更をステージング
git add .

# 5. コミット（meaningful なメッセージで）
git commit -m "feat: Add awesome feature"

# 6. プッシュ
git push origin main
```

### CI/CDで実行される内容

GitHub Actionsで自動的にチェックされる項目：

**test.yml（すべてのPR・Push）:**
- ✅ `go mod tidy` の整合性チェック
- ✅ `gofmt` でフォーマットチェック
- ✅ `go vet` で静的解析
- ✅ `go test` でテスト実行（race detector + coverage）
- ✅ `go build` でビルド確認
- ✅ `golangci-lint` で包括的なLint

**notify.yml（定期実行・手動実行）:**
- ✅ RSS取得と Discord通知の実行

### トラブルシューティング

#### CIでフォーマットエラーが出た場合

```bash
# エラー: 以下のファイルがフォーマットされていません
gofmt -s -w .
git add .
git commit -m "style: Format code with gofmt"
git push origin main
```

#### Lintエラーが出た場合

```bash
# エラー内容を確認
golangci-lint run ./...

# 修正してコミット
git add .
git commit -m "fix: Fix linter issues"
git push origin main
```

### 便利なMakefileコマンド一覧

| コマンド | 説明 |
|---------|------|
| `make help` | 利用可能なコマンドを表示 |
| `make all` | すべてのチェックとビルド |
| `make build` | バイナリをビルド |
| `make test` | テストを実行 |
| `make test-coverage` | カバレッジ付きテスト |
| `make lint` | Linterを実行 |
| `make fmt` | コードフォーマット |
| `make vet` | go vetを実行 |
| `make tidy` | go mod tidyを実行 |
| `make clean` | ビルド成果物を削除 |
| `make check` | フォーマット+vet+lint+test |
| `make deps` | 依存関係を表示 |
| `make deps-update` | 依存関係を最新化 |

---

## 📊 プロジェクト統計（2025-11-13時点）

### 実装済み機能

- ✅ RSS フィード取得（並行処理）
- ✅ Discord Webhook 通知（Embeds形式）
- ✅ **サムネイル画像表示**（RSSから画像を自動抽出）
- ✅ **フィードごとのチャンネル振り分け**（動的Webhook URL選択）
- ✅ 状態管理（通知済み記事の追跡）
- ✅ カテゴリ別の色分け
- ✅ 初回実行時の記事数制限
- ✅ エラーハンドリングとリトライ
- ✅ ログ出力（JSON/Text形式）
- ✅ GitHub Actions統合（Cron実行）
- ✅ CI/CD（テスト・Lint・ビルド）
- ✅ Dependabot（自動依存関係更新）
- ✅ **開発ワークフロー完備**（コミット前チェックリスト）

### コード品質

- **テストカバレッジ**: 55%
- **Goバージョン**: 1.21
- **Linter**: golangci-lint（10種類以上のLinter有効）
- **コード行数**: 約6,000行（ドキュメント含む）
- **パッケージ数**: 7個（models, config, feed, discord, state, logger, main）

### ドキュメント

- 📖 README.md（完全版）
- 📖 要件定義書
- 📖 アーキテクチャ設計書
- 📖 ADR（3件）
- 📖 Dependabotガイド
- 📖 AGENTS.md（開発記録）

