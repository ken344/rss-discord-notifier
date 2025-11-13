# ADR 0002: 状態管理方法の選定

## ステータス
承認済み

## コンテキスト
RSS Discord Notifierは同じ記事を重複して通知しないために、既に通知した記事の情報を保持する必要がある。GitHub Actionsの実行は状態を持たないため、実行間で状態を永続化する仕組みが必要となる。

### 要求事項
- 既に通知した記事を記録
- GitHub Actions実行間で状態を保持
- シンプルで保守しやすい
- 追加のインフラストラクチャを必要としない
- 無料で利用可能
- 複数のフィードに対応

### 検討した選択肢

#### 1. JSONファイル + GitHub Actions Cache
**概要**: 状態をJSONファイルに保存し、GitHub Actions Cacheで永続化

**メリット**:
- シンプルで実装が容易
- 追加のサービスや認証が不要
- GitHub Actionsに標準で組み込まれている
- ファイルベースで直感的

**デメリット**:
- Cacheは7日間で期限切れになる可能性
- Cacheのサイズ制限（10GB、通常は問題ない）
- Cacheがヒットしない場合、重複通知のリスク

**実装例**:
```yaml
- name: Cache state
  uses: actions/cache@v3
  with:
    path: state/
    key: rss-state-${{ github.run_id }}
    restore-keys: rss-state-
```

#### 2. JSONファイル + Gitコミット
**概要**: 状態ファイルをリポジトリにコミットして永続化

**メリット**:
- 確実に永続化される
- Git履歴で状態の変更を追跡可能
- Cacheの期限切れの心配がない

**デメリット**:
- 自動コミットのためのGitHub Token設定が必要
- コミット履歴が増える（ノイズになる可能性）
- ブランチ保護ルールとの整合性
- 複数の実行が並行した場合のコンフリクトリスク

**実装例**:
```yaml
- name: Commit state
  run: |
    git config user.name "github-actions[bot]"
    git config user.email "github-actions[bot]@users.noreply.github.com"
    git add state/state.json
    git commit -m "chore: update RSS state [skip ci]"
    git push
```

#### 3. GitHub Actions Artifacts
**概要**: 状態ファイルをArtifactとして保存し、次回実行時にダウンロード

**メリット**:
- 確実に保存される
- リポジトリを汚さない

**デメリット**:
- 90日で自動削除される
- 最新のArtifactを取得するロジックが複雑
- Artifact APIの利用が必要

#### 4. 外部ストレージ（S3、Redis等）
**概要**: 外部のストレージサービスに状態を保存

**メリット**:
- 確実で高可用性
- 複雑な状態管理にも対応可能

**デメリット**:
- 追加のインフラストラクチャが必要
- コスト（無料枠の範囲内でも設定が必要）
- 認証情報の管理
- プロジェクトの汎用性が低下

#### 5. RSSフィードの公開日時のみで判定
**概要**: 最後の実行時刻を保存し、それ以降の記事のみを通知

**メリット**:
- 最小限の状態（タイムスタンプのみ）

**デメリット**:
- RSSフィードの公開日時が正確でない場合に問題
- 記事の更新や再公開に対応できない
- 信頼性が低い

## 決定
**JSONファイル + GitHub Actions Cache（メイン） + Artifacts（バックアップ）のハイブリッド方式**

具体的には：
1. 状態を`state/state.json`に保存
2. GitHub Actions Cacheでキャッシュをメインとする
3. Artifactsにもバックアップとしてアップロード
4. Cacheが期限切れの場合、Artifactsからリストア

## 理由

### 1. シンプルさとバランス
Cacheをメインにすることで、通常の運用では高速で簡単。Artifactsをバックアップにすることで、Cacheが期限切れの場合にも対応可能。

### 2. 追加インフラ不要
GitHub Actionsの標準機能のみで完結し、他のサービスのアカウントや認証情報が不要。プロジェクトのクローン性が保たれる。

### 3. リポジトリを汚さない
自動コミットによるノイズを避けられる。状態管理はランタイムの関心事であり、コード管理とは分離すべき。

### 4. 柔軟性
将来的に状態管理を外部サービスに移行する場合も、インターフェースを変更するだけで対応可能。

## 実装詳細

### 状態ファイル形式（state.json）

```json
{
  "version": "1.0",
  "last_update": "2025-11-13T10:00:00Z",
  "feeds": {
    "https://go.dev/blog/feed.atom": {
      "last_check": "2025-11-13T10:00:00Z",
      "articles": [
        {
          "id": "https://go.dev/blog/article-123",
          "title": "Go 1.22 Released",
          "url": "https://go.dev/blog/article-123",
          "published_at": "2025-11-13T09:00:00Z",
          "notified_at": "2025-11-13T10:00:00Z"
        }
      ]
    }
  },
  "statistics": {
    "total_articles_notified": 42,
    "total_feeds_checked": 5
  }
}
```

### データ構造（Go）

```go
type State struct {
    Version    string                `json:"version"`
    LastUpdate time.Time             `json:"last_update"`
    Feeds      map[string]FeedState  `json:"feeds"`
    Statistics Statistics            `json:"statistics"`
}

type FeedState struct {
    LastCheck time.Time        `json:"last_check"`
    Articles  []NotifiedArticle `json:"articles"`
}

type NotifiedArticle struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    URL         string    `json:"url"`
    PublishedAt time.Time `json:"published_at"`
    NotifiedAt  time.Time `json:"notified_at"`
}

type Statistics struct {
    TotalArticlesNotified int `json:"total_articles_notified"`
    TotalFeedsChecked     int `json:"total_feeds_checked"`
}
```

### GitHub Actions ワークフロー

```yaml
jobs:
  notify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      # メイン: Cache
      - name: Restore state from cache
        id: cache-state
        uses: actions/cache/restore@v3
        with:
          path: state/
          key: rss-state-${{ github.run_number }}
          restore-keys: rss-state-
      
      # バックアップ: Artifacts（Cacheがない場合）
      - name: Download state from artifacts
        if: steps.cache-state.outputs.cache-hit != 'true'
        uses: dawidd6/action-download-artifact@v2
        continue-on-error: true
        with:
          name: rss-state
          path: state/
          workflow_conclusion: success
      
      - name: Run notifier
        env:
          DISCORD_WEBHOOK_URL: ${{ secrets.DISCORD_WEBHOOK_URL }}
        run: go run cmd/notifier/main.go
      
      # 状態の保存: Cache
      - name: Save state to cache
        uses: actions/cache/save@v3
        if: always()
        with:
          path: state/
          key: rss-state-${{ github.run_number }}
      
      # 状態の保存: Artifacts
      - name: Upload state to artifacts
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: rss-state
          path: state/
          retention-days: 30
```

### 状態のクリーンアップ戦略
- 30日以上前の記事情報は削除（メモリとファイルサイズの最適化）
- フィードごとに最新100件のみ保持
- 無効化されたフィードの状態は次回実行時に削除

## 影響

### ポジティブ
- ✅ シンプルで理解しやすい
- ✅ 追加コストなし
- ✅ 高い汎用性（誰でもクローンして使える）
- ✅ Cache + Artifactsの2重化で信頼性向上

### ネガティブ
- ⚠️ Cacheが7日間でクリアされる可能性
- ⚠️ Artifactsのダウンロードロジックがやや複雑
- ⚠️ 長期間実行しない場合、再通知のリスク

### リスク軽減策
- 初回実行時は最新N件のみを通知（N=5など）
- 状態ファイルがない場合の挙動を明確に定義
- ログで状態の読み込み成否を記録

## 代替案（将来的な拡張）
プロジェクトが成長した場合、以下の選択肢を検討可能：
- GitHub Gist（APIで読み書き）
- Cloudflare KV（無料枠あり）
- GitHub GraphQL API（Issueやディスカッションに状態を保存）

## 関連決定
- ADR 0001: Go言語の採用

## 参考
- [GitHub Actions Cache](https://docs.github.com/en/actions/using-workflows/caching-dependencies-to-speed-up-workflows)
- [GitHub Actions Artifacts](https://docs.github.com/en/actions/using-workflows/storing-workflow-data-as-artifacts)

