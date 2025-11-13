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

