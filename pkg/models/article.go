package models

import "time"

// Article は、RSSフィードから取得した記事を表すモデル
type Article struct {
	// ID は記事の一意な識別子（通常はGUIDまたはURL）
	ID string

	// Title は記事のタイトル
	Title string

	// URL は記事のURL
	URL string

	// Description は記事の説明文または要約
	Description string

	// Content は記事の本文（RSSフィードによっては空の場合がある）
	Content string

	// Author は記事の著者名
	Author string

	// PublishedAt は記事の公開日時
	PublishedAt time.Time

	// UpdatedAt は記事の更新日時
	UpdatedAt time.Time

	// FeedName はこの記事が属するフィード名
	FeedName string

	// FeedURL はこの記事が属するフィードのURL
	FeedURL string

	// Category はフィードのカテゴリ（Tech, News, Blog, Otherなど）
	Category string

	// WebhookURL はこの記事を通知する際に使用するDiscord Webhook URL
	// フィード設定から引き継がれる
	WebhookURL string
}

// IsValid は、記事が有効なデータを持っているかチェックする
func (a *Article) IsValid() bool {
	// 最低限、IDとタイトルとURLが必要
	return a.ID != "" && a.Title != "" && a.URL != ""
}

// GetShortDescription は、説明文を指定された長さに切り詰める
func (a *Article) GetShortDescription(maxLength int) string {
	if len(a.Description) <= maxLength {
		return a.Description
	}
	
	// maxLengthを超える場合は切り詰めて "..." を追加
	if maxLength > 3 {
		return a.Description[:maxLength-3] + "..."
	}
	
	return a.Description[:maxLength]
}

