# ADR 0003: Discord通知フォーマットの決定

## ステータス
承認済み

## コンテキスト
RSSフィードから取得した記事をDiscordに通知する際、どのようなフォーマットで情報を提示するかを決定する必要がある。Discordは複数の通知形式をサポートしており、視認性、情報量、使いやすさのバランスを考慮する必要がある。

### 要求事項
- 記事の重要情報（タイトル、URL、要約、公開日時）を含む
- 視認性が高く、一目で内容を把握できる
- 複数の記事が連続して投稿される場合も読みやすい
- Discord Webhookで実装可能
- モバイルでも見やすい

### 検討した選択肢

#### 1. プレーンテキスト
**概要**: シンプルなテキストメッセージ

**例**:
```
[Go Blog] Go 1.22 Released
https://go.dev/blog/go1.22
公開: 2025-11-13 09:00

Go 1.22がリリースされました。新機能として...
```

**メリット**:
- シンプルで実装が容易
- メッセージサイズが小さい

**デメリット**:
- 視認性が低い
- リンクプレビューが表示されない場合がある
- 複数記事が並ぶと見づらい
- カテゴリやメタ情報の構造化が困難

#### 2. Discord Embeds（単一フィールド）
**概要**: Embedを使用し、最小限の情報のみ

**メリット**:
- 視覚的に区別しやすい
- クリック可能なタイトルリンク

**デメリット**:
- 情報量が少ない
- カスタマイズ性が低い

#### 3. Discord Embeds（リッチフォーマット）
**概要**: Embedの全機能を活用し、構造化された情報を提供

**メリット**:
- 高い視認性と美しい見た目
- 情報が構造化されている
- フィールドでメタ情報を整理
- カラーコーディング可能
- サムネイルやフッター情報も追加可能

**デメリット**:
- 実装がやや複雑
- メッセージサイズが大きい（Embedは1つのメッセージに最大10個まで）

#### 4. Discord Embeds + ボタン
**概要**: Embedに加えてインタラクティブなボタンを追加

**メリット**:
- ユーザーアクション（既読、ブックマーク等）を組み込める

**デメリット**:
- WebhookではInteractionに対応できない（Discord Botが必要）
- プロジェクトのスコープを超える

## 決定
**Discord Embeds（リッチフォーマット）を採用する**

## 理由

### 1. 視認性と情報構造
Embedsを使用することで、記事の各要素（タイトル、URL、要約、メタ情報）を明確に区別でき、ユーザーは一目で必要な情報を把握できる。

### 2. Discord UIとの親和性
Embedsは Discord の標準的な情報提示方法であり、ユーザーが慣れ親しんだフォーマット。

### 3. カスタマイズ性
将来的にフィードごとに色を変更したり、サムネイル画像を追加したりする拡張が容易。

### 4. モバイル対応
Discord のモバイルアプリでもEmbedsは適切にレンダリングされる。

## 実装詳細

### 基本フォーマット

```json
{
  "embeds": [{
    "title": "記事タイトル",
    "url": "https://example.com/article",
    "description": "記事の要約または冒頭200文字",
    "color": 5814783,
    "fields": [
      {
        "name": "📰 フィード",
        "value": "Go Blog",
        "inline": true
      },
      {
        "name": "📅 公開日時",
        "value": "2025-11-13 09:00",
        "inline": true
      },
      {
        "name": "🏷️ カテゴリ",
        "value": "Tech",
        "inline": true
      }
    ],
    "timestamp": "2025-11-13T10:00:00Z",
    "footer": {
      "text": "RSS Discord Notifier"
    }
  }]
}
```

### カラーコーディング

フィードのカテゴリごとに色を変更し、視覚的に区別しやすくする：

| カテゴリ | 色コード（16進数） | 色コード（10進数） | 色 |
|---------|-----------------|-------------------|-----|
| Tech    | `#5865F2`       | 5793522          | Discord Blurple |
| News    | `#57F287`       | 5763719          | Green |
| Blog    | `#FEE75C`       | 16770908         | Yellow |
| その他  | `#EB459E`       | 15418782         | Pink |

### フィールド構成

1. **タイトル**: クリック可能なリンク（最大256文字）
2. **説明**: 記事の要約（最大4096文字、実際には200-300文字に制限）
3. **フィールド**:
   - フィード名（inline）
   - 公開日時（inline）
   - カテゴリ（inline）
4. **タイムスタンプ**: 記事の公開日時（ISO 8601形式）
5. **フッター**: アプリケーション名

### Go実装（サンプル）

```go
type DiscordEmbed struct {
    Title       string         `json:"title"`
    URL         string         `json:"url"`
    Description string         `json:"description"`
    Color       int            `json:"color"`
    Fields      []EmbedField   `json:"fields"`
    Timestamp   string         `json:"timestamp"`
    Footer      *EmbedFooter   `json:"footer,omitempty"`
}

type EmbedField struct {
    Name   string `json:"name"`
    Value  string `json:"value"`
    Inline bool   `json:"inline"`
}

type EmbedFooter struct {
    Text string `json:"text"`
}

type WebhookMessage struct {
    Embeds []DiscordEmbed `json:"embeds"`
}

func CreateArticleEmbed(article Article, feedName, category string) DiscordEmbed {
    description := truncateText(article.Description, 300)
    color := getCategoryColor(category)
    
    return DiscordEmbed{
        Title:       article.Title,
        URL:         article.URL,
        Description: description,
        Color:       color,
        Fields: []EmbedField{
            {
                Name:   "📰 フィード",
                Value:  feedName,
                Inline: true,
            },
            {
                Name:   "📅 公開日時",
                Value:  article.PublishedAt.Format("2006-01-02 15:04"),
                Inline: true,
            },
            {
                Name:   "🏷️ カテゴリ",
                Value:  category,
                Inline: true,
            },
        },
        Timestamp: article.PublishedAt.Format(time.RFC3339),
        Footer: &EmbedFooter{
            Text: "RSS Discord Notifier",
        },
    }
}

func getCategoryColor(category string) int {
    colors := map[string]int{
        "Tech":  5793522,  // Discord Blurple
        "News":  5763719,  // Green
        "Blog":  16770908, // Yellow
        "Other": 15418782, // Pink
    }
    
    if color, ok := colors[category]; ok {
        return color
    }
    return 5793522 // デフォルト: Discord Blurple
}

func truncateText(text string, maxLength int) string {
    if len(text) <= maxLength {
        return text
    }
    return text[:maxLength-3] + "..."
}
```

### バッチ送信の考慮

Discord Webhookは1つのメッセージに最大10個のEmbedsを含めることができる。複数記事を通知する場合：

**オプションA: 個別メッセージ**（推奨）
- 各記事を個別のメッセージとして送信
- レート制限を考慮して間隔を空ける（1秒など）
- 通知の順序が保証される
- 個別の通知として認識されやすい

**オプションB: バッチメッセージ**
- 最大10記事を1つのメッセージにまとめる
- API呼び出し数を削減
- 通知がまとまりすぎて見づらくなる可能性

→ **オプションAを採用**: ユーザー体験を優先

### レート制限対策

Discord Webhookのレート制限：
- 5リクエスト/秒
- 30リクエスト/分

実装では：
```go
const (
    RateLimitDelay = 1 * time.Second  // 各メッセージ間の遅延
    MaxRetries     = 3                // リトライ回数
    RetryDelay     = 5 * time.Second  // リトライ間隔
)

func (n *Notifier) SendArticles(ctx context.Context, articles []Article) error {
    for _, article := range articles {
        if err := n.SendArticle(ctx, article); err != nil {
            log.Error("Failed to send article", "error", err, "article", article.Title)
            continue // エラーがあっても次の記事は送信を試みる
        }
        
        // レート制限対策: 各送信後に待機
        time.Sleep(RateLimitDelay)
    }
    return nil
}
```

## 影響

### ポジティブ
- ✅ 高い視認性と美しい見た目
- ✅ 情報が構造化され、理解しやすい
- ✅ カテゴリごとの色分けで直感的
- ✅ モバイルでも見やすい
- ✅ 将来的な拡張が容易

### ネガティブ
- ⚠️ プレーンテキストよりもペイロードサイズが大きい
- ⚠️ 実装がやや複雑

### トレードオフ
視認性とユーザー体験を優先し、実装の複雑さは許容する。

## 制約事項

### Discord Embedsの制限
- タイトル: 最大256文字
- 説明: 最大4096文字（実際には300文字程度に制限）
- フィールド数: 最大25個
- フィールド名: 最大256文字
- フィールド値: 最大1024文字
- フッター: 最大2048文字
- Embed総文字数: 最大6000文字
- 1メッセージあたりのEmbeds: 最大10個

## 実装された拡張機能

### 1. サムネイル画像 ✅ 実装済み（2025-11-13）
記事にOGP画像がある場合、サムネイルとして表示：
```json
"thumbnail": {
  "url": "https://example.com/image.jpg"
}
```

**実装内容:**
- RSSフィードの`<image>`タグから画像URLを自動抽出
- `<enclosure type="image/*">`からも画像を取得
- Discord EmbedのThumbnailフィールドに設定
- 画像がない記事でもエラーにならず正常動作

**実装ファイル:**
- `pkg/models/article.go`: `ImageURL`フィールド追加
- `internal/feed/fetcher.go`: `extractImageURL()`メソッド実装
- `internal/discord/notifier.go`: Thumbnail設定ロジック追加

## 将来的な拡張案

### 1. 著者アイコン画像
現在、著者名は表示されていますが、アイコン画像は未対応：
```json
"author": {
  "name": "著者名",
  "icon_url": "https://example.com/avatar.jpg"
}
```

### 2. カスタマイズ可能なテンプレート
設定ファイルでEmbed形式をカスタマイズ可能にする：
```yaml
notification_template:
  show_category: true
  show_published_date: true
  show_footer: true
  color_by_category: true
```

## 関連決定
- ADR 0001: Go言語の採用
- ADR 0002: 状態管理方法

## 参考
- [Discord Webhook Documentation](https://discord.com/developers/docs/resources/webhook)
- [Discord Embed Visualizer](https://discohook.org/)
- [Discord Embed Limits](https://discord.com/developers/docs/resources/channel#embed-object-embed-limits)

