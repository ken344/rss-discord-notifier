package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ken344/rss-discord-notifier/internal/logger"
	"github.com/ken344/rss-discord-notifier/pkg/models"
)

// Notifier は、Discordに通知を送信する構造体
type Notifier struct {
	// webhookURL はDiscord Webhook URL
	webhookURL string

	// client はHTTPクライアント
	client *http.Client

	// rateLimit は通知間隔（レート制限対策）
	rateLimit time.Duration

	// maxRetries は最大リトライ回数
	maxRetries int

	// retryDelay はリトライ間隔
	retryDelay time.Duration
}

// NewNotifier は、新しいDiscord通知器を作成する
func NewNotifier(webhookURL string, rateLimit time.Duration) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimit:  rateLimit,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}
}

// SendArticle は、単一の記事をDiscordに通知する
func (n *Notifier) SendArticle(ctx context.Context, article *models.Article) error {
	// Embedメッセージを作成
	message := n.createMessage(article)

	// 送信（リトライ付き）
	err := n.sendWithRetry(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send article %s: %w", article.Title, err)
	}

	logger.Info("記事を通知しました",
		"title", article.Title,
		"feed", article.FeedName,
		"category", article.Category)

	return nil
}

// SendArticles は、複数の記事をDiscordに通知する
func (n *Notifier) SendArticles(ctx context.Context, articles []*models.Article) error {
	if len(articles) == 0 {
		logger.Info("通知する記事がありません")
		return nil
	}

	logger.Info("記事の通知を開始", "count", len(articles))
	successCount := 0
	errorCount := 0

	for i, article := range articles {
		// コンテキストがキャンセルされたかチェック
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 記事を送信
		if err := n.SendArticle(ctx, article); err != nil {
			logger.Error("記事の通知に失敗",
				"title", article.Title,
				"error", err)
			errorCount++
			continue
		}

		successCount++

		// 最後の記事以外は、レート制限対策で待機
		if i < len(articles)-1 {
			select {
			case <-time.After(n.rateLimit):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	logger.Info("記事の通知が完了",
		"total", len(articles),
		"success", successCount,
		"failed", errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to send %d articles", errorCount)
	}

	return nil
}

// createMessage は、記事からDiscordメッセージを作成する
func (n *Notifier) createMessage(article *models.Article) *WebhookMessage {
	// 説明文を短縮（最大300文字）
	description := article.GetShortDescription(300)

	// カテゴリに応じた色を取得
	color := getCategoryColor(article.Category)

	// Embedを作成
	embed := Embed{
		Title:       article.Title,
		URL:         article.URL,
		Description: description,
		Color:       color,
		Fields: []EmbedField{
			{
				Name:   "📰 フィード",
				Value:  article.FeedName,
				Inline: true,
			},
			{
				Name:   "📅 公開日時",
				Value:  article.PublishedAt.Format("2006-01-02 15:04"),
				Inline: true,
			},
			{
				Name:   "🏷️ カテゴリ",
				Value:  article.Category,
				Inline: true,
			},
		},
		Timestamp: article.PublishedAt.Format(time.RFC3339),
		Footer: &EmbedFooter{
			Text: "RSS Discord Notifier",
		},
	}

	// 著者情報があれば追加
	if article.Author != "" {
		embed.Author = &EmbedAuthor{
			Name: article.Author,
		}
	}

	// 画像URLがあればサムネイルとして追加
	if article.ImageURL != "" {
		embed.Thumbnail = &EmbedImage{
			URL: article.ImageURL,
		}
	}

	return &WebhookMessage{
		Embeds: []Embed{embed},
	}
}

// sendWithRetry は、リトライ付きでメッセージを送信する
func (n *Notifier) sendWithRetry(ctx context.Context, message *WebhookMessage) error {
	var lastErr error

	for attempt := 0; attempt < n.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("Discord送信をリトライ", "attempt", attempt+1)
			select {
			case <-time.After(n.retryDelay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := n.send(ctx, message)
		if err == nil {
			return nil
		}

		lastErr = err
		logger.Warn("Discord送信に失敗",
			"attempt", attempt+1,
			"max_retries", n.maxRetries,
			"error", err)
	}

	return fmt.Errorf("failed after %d retries: %w", n.maxRetries, lastErr)
}

// send は、メッセージをDiscordに送信する
func (n *Notifier) send(ctx context.Context, message *WebhookMessage) error {
	// JSONにエンコード
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	logger.Debug("Discord Webhookに送信中", "url", n.webhookURL)

	// HTTPリクエストを作成
	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエスト送信
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord API returned error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	logger.Debug("Discord Webhookへの送信が成功", "status", resp.StatusCode)

	return nil
}

// getCategoryColor は、カテゴリに応じた色コードを返す
func getCategoryColor(category string) int {
	colors := map[string]int{
		"Tech":  5793522,  // Discord Blurple (#5865F2)
		"News":  5763719,  // Green (#57F287)
		"Blog":  16770908, // Yellow (#FEE75C)
		"Other": 15418782, // Pink (#EB459E)
	}

	if color, ok := colors[category]; ok {
		return color
	}

	// デフォルト: Discord Blurple
	return 5793522
}

// SetRateLimit は、レート制限間隔を設定する
func (n *Notifier) SetRateLimit(duration time.Duration) {
	n.rateLimit = duration
}

// SetMaxRetries は、最大リトライ回数を設定する
func (n *Notifier) SetMaxRetries(count int) {
	if count > 0 {
		n.maxRetries = count
	}
}

// SetRetryDelay は、リトライ間隔を設定する
func (n *Notifier) SetRetryDelay(duration time.Duration) {
	n.retryDelay = duration
}
