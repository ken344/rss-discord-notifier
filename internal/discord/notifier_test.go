package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ken344/rss-discord-notifier/internal/logger"
	"github.com/ken344/rss-discord-notifier/pkg/models"
)

func init() {
	// テスト用のロガーを初期化（出力を抑制）
	logger.Init(&logger.Config{
		Level:  logger.LevelError,
		Format: logger.FormatJSON,
		Output: os.Stderr,
	})
}

// TestNewNotifier は、Notifierの作成をテストする
func TestNewNotifier(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/test"
	rateLimit := 1 * time.Second

	notifier := NewNotifier(webhookURL, rateLimit)

	if notifier == nil {
		t.Fatal("NewNotifier() should not return nil")
	}

	if notifier.webhookURL != webhookURL {
		t.Errorf("webhookURL = %v, want %v", notifier.webhookURL, webhookURL)
	}

	if notifier.rateLimit != rateLimit {
		t.Errorf("rateLimit = %v, want %v", notifier.rateLimit, rateLimit)
	}

	if notifier.client == nil {
		t.Error("client should not be nil")
	}
}

// TestCreateMessage は、メッセージ作成をテストする
func TestCreateMessage(t *testing.T) {
	notifier := NewNotifier("https://test.com", 1*time.Second)

	publishedAt := time.Date(2025, 11, 13, 10, 0, 0, 0, time.UTC)

	article := &models.Article{
		ID:          "article-1",
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		Description: "This is a test article description.",
		Author:      "Test Author",
		PublishedAt: publishedAt,
		FeedName:    "Test Feed",
		FeedURL:     "https://example.com/feed",
		Category:    "Tech",
	}

	message := notifier.createMessage(article)

	if message == nil {
		t.Fatal("createMessage() should not return nil")
	}

	if len(message.Embeds) != 1 {
		t.Fatalf("Embeds length = %d, want 1", len(message.Embeds))
	}

	embed := message.Embeds[0]

	// タイトルのチェック
	if embed.Title != article.Title {
		t.Errorf("Title = %v, want %v", embed.Title, article.Title)
	}

	// URLのチェック
	if embed.URL != article.URL {
		t.Errorf("URL = %v, want %v", embed.URL, article.URL)
	}

	// 色のチェック（Tech = 5793522）
	if embed.Color != 5793522 {
		t.Errorf("Color = %d, want 5793522", embed.Color)
	}

	// フィールド数のチェック
	if len(embed.Fields) != 3 {
		t.Errorf("Fields length = %d, want 3", len(embed.Fields))
	}

	// 著者情報のチェック
	if embed.Author == nil {
		t.Error("Author should not be nil")
	} else if embed.Author.Name != article.Author {
		t.Errorf("Author.Name = %v, want %v", embed.Author.Name, article.Author)
	}

	// フッターのチェック
	if embed.Footer == nil {
		t.Error("Footer should not be nil")
	}
}

// TestGetCategoryColor は、カテゴリ別色分けをテストする
func TestGetCategoryColor(t *testing.T) {
	tests := []struct {
		category string
		want     int
	}{
		{"Tech", 5793522},
		{"News", 5763719},
		{"Blog", 16770908},
		{"Other", 15418782},
		{"Unknown", 5793522}, // デフォルト
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := getCategoryColor(tt.category)
			if got != tt.want {
				t.Errorf("getCategoryColor(%q) = %d, want %d", tt.category, got, tt.want)
			}
		})
	}
}

// TestSendArticle は、記事送信をテストする（モックサーバー使用）
func TestSendArticle(t *testing.T) {
	// モックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// リクエストメソッドをチェック
		if r.Method != "POST" {
			t.Errorf("Method = %v, want POST", r.Method)
		}

		// Content-Typeをチェック
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %v, want application/json", ct)
		}

		// リクエストボディをパース
		var message WebhookMessage
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// レスポンスを返す
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// Notifierを作成（モックサーバーのURLを使用）
	notifier := NewNotifier(server.URL, 100*time.Millisecond)
	notifier.SetMaxRetries(1) // テストを高速化

	ctx := context.Background()
	article := &models.Article{
		ID:          "article-1",
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		Description: "Test description",
		PublishedAt: time.Now(),
		FeedName:    "Test Feed",
		FeedURL:     "https://example.com/feed",
		Category:    "Tech",
	}

	err := notifier.SendArticle(ctx, article)
	if err != nil {
		t.Errorf("SendArticle() error = %v, want nil", err)
	}
}

// TestSendArticles は、複数記事送信をテストする
func TestSendArticles(t *testing.T) {
	requestCount := 0

	// モックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewNotifier(server.URL, 10*time.Millisecond)
	notifier.SetMaxRetries(1)

	ctx := context.Background()
	articles := []*models.Article{
		{
			ID:          "article-1",
			Title:       "Article 1",
			URL:         "https://example.com/article-1",
			PublishedAt: time.Now(),
			FeedName:    "Test Feed",
			FeedURL:     "https://example.com/feed",
			Category:    "Tech",
		},
		{
			ID:          "article-2",
			Title:       "Article 2",
			URL:         "https://example.com/article-2",
			PublishedAt: time.Now(),
			FeedName:    "Test Feed",
			FeedURL:     "https://example.com/feed",
			Category:    "News",
		},
	}

	err := notifier.SendArticles(ctx, articles)
	if err != nil {
		t.Errorf("SendArticles() error = %v, want nil", err)
	}

	// リクエスト数をチェック
	if requestCount != len(articles) {
		t.Errorf("Request count = %d, want %d", requestCount, len(articles))
	}
}

// TestSendWithRetry は、リトライ処理をテストする
func TestSendWithRetry(t *testing.T) {
	attemptCount := 0
	maxAttempts := 2

	// モックサーバー（最初は失敗、2回目は成功）
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < maxAttempts {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewNotifier(server.URL, 100*time.Millisecond)
	notifier.SetMaxRetries(3)
	notifier.SetRetryDelay(10 * time.Millisecond) // テストを高速化

	ctx := context.Background()
	article := &models.Article{
		ID:          "article-1",
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		PublishedAt: time.Now(),
		FeedName:    "Test Feed",
		FeedURL:     "https://example.com/feed",
		Category:    "Tech",
	}

	err := notifier.SendArticle(ctx, article)
	if err != nil {
		t.Errorf("SendArticle() error = %v, want nil (should succeed after retry)", err)
	}

	if attemptCount != maxAttempts {
		t.Errorf("Attempt count = %d, want %d", attemptCount, maxAttempts)
	}
}

// TestSendArticlesEmpty は、空の記事リストの処理をテストする
func TestSendArticlesEmpty(t *testing.T) {
	notifier := NewNotifier("https://test.com", 1*time.Second)
	ctx := context.Background()

	err := notifier.SendArticles(ctx, []*models.Article{})
	if err != nil {
		t.Errorf("SendArticles() with empty list should not return error, got %v", err)
	}
}

// TestSetters は、セッターメソッドをテストする
func TestSetters(t *testing.T) {
	notifier := NewNotifier("https://test.com", 1*time.Second)

	// SetRateLimit
	newRateLimit := 2 * time.Second
	notifier.SetRateLimit(newRateLimit)
	if notifier.rateLimit != newRateLimit {
		t.Errorf("rateLimit = %v, want %v", notifier.rateLimit, newRateLimit)
	}

	// SetMaxRetries
	newMaxRetries := 5
	notifier.SetMaxRetries(newMaxRetries)
	if notifier.maxRetries != newMaxRetries {
		t.Errorf("maxRetries = %d, want %d", notifier.maxRetries, newMaxRetries)
	}

	// SetRetryDelay
	newRetryDelay := 10 * time.Second
	notifier.SetRetryDelay(newRetryDelay)
	if notifier.retryDelay != newRetryDelay {
		t.Errorf("retryDelay = %v, want %v", notifier.retryDelay, newRetryDelay)
	}
}
