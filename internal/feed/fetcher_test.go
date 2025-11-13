package feed

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ken344/rss-discord-notifier/internal/logger"
	"github.com/ken344/rss-discord-notifier/pkg/models"
	"github.com/mmcdole/gofeed"
)

func init() {
	// テスト用のロガーを初期化（出力を抑制）
	logger.Init(&logger.Config{
		Level:  logger.LevelError,
		Format: logger.FormatJSON,
		Output: os.Stderr,
	})
}

// TestNewFetcher は、フィード取得器の作成をテストする
func TestNewFetcher(t *testing.T) {
	timeout := 30 * time.Second
	fetcher := NewFetcher(timeout)

	if fetcher == nil {
		t.Fatal("NewFetcher() should not return nil")
	}

	if fetcher.parser == nil {
		t.Error("parser should not be nil")
	}

	if fetcher.timeout != timeout {
		t.Errorf("timeout = %v, want %v", fetcher.timeout, timeout)
	}
}

// TestConvertToArticle は、記事変換をテストする
func TestConvertToArticle(t *testing.T) {
	fetcher := NewFetcher(30 * time.Second)

	publishedTime := time.Now()
	updatedTime := time.Now().Add(1 * time.Hour)

	feedConfig := &models.FeedConfig{
		Name:     "Test Feed",
		URL:      "https://example.com/feed",
		Category: "Tech",
		Enabled:  true,
	}

	feed := &gofeed.Feed{
		Title:       "Test Feed Title",
		Description: "Test Feed Description",
	}

	tests := []struct {
		name string
		item *gofeed.Item
		want *models.Article
	}{
		{
			name: "完全な記事",
			item: &gofeed.Item{
				GUID:            "article-1",
				Title:           "Test Article",
				Link:            "https://example.com/article-1",
				Description:     "Test Description",
				Content:         "Test Content",
				Author:          &gofeed.Person{Name: "Test Author"},
				PublishedParsed: &publishedTime,
				UpdatedParsed:   &updatedTime,
			},
			want: &models.Article{
				ID:          "article-1",
				Title:       "Test Article",
				URL:         "https://example.com/article-1",
				Description: "Test Description",
				Content:     "Test Content",
				Author:      "Test Author",
				PublishedAt: publishedTime,
				UpdatedAt:   updatedTime,
				FeedName:    "Test Feed",
				FeedURL:     "https://example.com/feed",
				Category:    "Tech",
			},
		},
		{
			name: "GUIDがない場合（LinkをIDに使用）",
			item: &gofeed.Item{
				GUID:            "",
				Title:           "Test Article 2",
				Link:            "https://example.com/article-2",
				Description:     "Description 2",
				PublishedParsed: &publishedTime,
			},
			want: &models.Article{
				ID:          "https://example.com/article-2",
				Title:       "Test Article 2",
				URL:         "https://example.com/article-2",
				Description: "Description 2",
				FeedName:    "Test Feed",
				FeedURL:     "https://example.com/feed",
				Category:    "Tech",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fetcher.convertToArticle(tt.item, feedConfig, feed)

			if got == nil {
				t.Fatal("convertToArticle() returned nil")
			}

			// 主要フィールドをチェック
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Title != tt.want.Title {
				t.Errorf("Title = %v, want %v", got.Title, tt.want.Title)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %v, want %v", got.URL, tt.want.URL)
			}
			if got.FeedName != tt.want.FeedName {
				t.Errorf("FeedName = %v, want %v", got.FeedName, tt.want.FeedName)
			}
			if got.Category != tt.want.Category {
				t.Errorf("Category = %v, want %v", got.Category, tt.want.Category)
			}
		})
	}
}

// TestStripHTML は、HTMLタグ除去をテストする
func TestStripHTML(t *testing.T) {
	fetcher := NewFetcher(30 * time.Second)

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "タグなし",
			input: "Plain text",
			want:  "Plain text",
		},
		{
			name:  "シンプルなタグ",
			input: "<p>Hello <strong>World</strong></p>",
			want:  "Hello World",
		},
		{
			name:  "改行タグ",
			input: "Line 1<br>Line 2<br/>Line 3",
			want:  "Line 1 Line 2 Line 3",
		},
		{
			name:  "段落タグ",
			input: "<p>Paragraph 1</p><p>Paragraph 2</p>",
			want:  "Paragraph 1 Paragraph 2",
		},
		{
			name:  "複雑なHTML",
			input: "<div><h1>Title</h1><p>Content with <a href='#'>link</a></p></div>",
			want:  "Title Content with link",
		},
		{
			name:  "空文字列",
			input: "",
			want:  "",
		},
		{
			name:  "連続する空白",
			input: "Multiple    spaces    here",
			want:  "Multiple spaces here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fetcher.stripHTML(tt.input)
			if got != tt.want {
				t.Errorf("stripHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFetchAll は、複数フィードの並行取得をテストする
func TestFetchAll(t *testing.T) {
	fetcher := NewFetcher(30 * time.Second)
	ctx := context.Background()

	tests := []struct {
		name        string
		feedConfigs []*models.FeedConfig
		wantErr     bool
	}{
		{
			name:        "空のフィードリスト",
			feedConfigs: []*models.FeedConfig{},
			wantErr:     false,
		},
		{
			name: "無効なURL",
			feedConfigs: []*models.FeedConfig{
				{
					Name:     "Invalid Feed",
					URL:      "invalid-url",
					Category: "Tech",
					Enabled:  true,
				},
			},
			wantErr: false, // エラーがあっても処理は継続
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			articles, err := fetcher.FetchAll(ctx, tt.feedConfigs)
			if (err != nil) != tt.wantErr {
				t.Errorf("FetchAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 空のフィードリストの場合、記事も空
			if len(tt.feedConfigs) == 0 && len(articles) != 0 {
				t.Errorf("FetchAll() with empty config should return empty articles, got %d", len(articles))
			}
		})
	}
}

// TestFetch は、単一フィードの取得をテストする
func TestFetch(t *testing.T) {
	fetcher := NewFetcher(30 * time.Second)
	ctx := context.Background()

	tests := []struct {
		name       string
		feedConfig *models.FeedConfig
		wantErr    bool
	}{
		{
			name: "無効なURL",
			feedConfig: &models.FeedConfig{
				Name:     "Invalid Feed",
				URL:      "invalid-url",
				Category: "Tech",
				Enabled:  true,
			},
			wantErr: true,
		},
		{
			name: "存在しないURL",
			feedConfig: &models.FeedConfig{
				Name:     "Not Found Feed",
				URL:      "https://example.com/nonexistent-feed.xml",
				Category: "Tech",
				Enabled:  true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fetcher.Fetch(ctx, tt.feedConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFetchWithTimeout は、タイムアウト処理をテストする
func TestFetchWithTimeout(t *testing.T) {
	// 極端に短いタイムアウトを設定
	fetcher := NewFetcher(1 * time.Nanosecond)
	ctx := context.Background()

	feedConfig := &models.FeedConfig{
		Name:     "Timeout Feed",
		URL:      "https://example.com/feed.xml",
		Category: "Tech",
		Enabled:  true,
	}

	_, err := fetcher.Fetch(ctx, feedConfig)
	// タイムアウトまたはその他のエラーが発生するはず
	if err == nil {
		t.Error("Fetch() with very short timeout should return error")
	}
}
