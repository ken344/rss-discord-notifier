package feed

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ken344/rss-discord-notifier/internal/logger"
	"github.com/ken344/rss-discord-notifier/pkg/models"
	"github.com/mmcdole/gofeed"
)

// Fetcher は、RSSフィードを取得する構造体
type Fetcher struct {
	// parser はRSSパーサー
	parser *gofeed.Parser

	// timeout はフィード取得のタイムアウト
	timeout time.Duration
}

// NewFetcher は、新しいフィード取得器を作成する
func NewFetcher(timeout time.Duration) *Fetcher {
	return &Fetcher{
		parser:  gofeed.NewParser(),
		timeout: timeout,
	}
}

// FetchAll は、複数のフィードから記事を並行で取得する
func (f *Fetcher) FetchAll(ctx context.Context, feedConfigs []*models.FeedConfig) ([]*models.Article, error) {
	if len(feedConfigs) == 0 {
		return []*models.Article{}, nil
	}

	logger.Info("RSSフィードの取得を開始", "feed_count", len(feedConfigs))
	startTime := time.Now()

	// 結果を格納するチャネル
	type result struct {
		articles []*models.Article
		err      error
		feedName string
	}
	resultChan := make(chan result, len(feedConfigs))

	// 各フィードを並行で取得
	var wg sync.WaitGroup
	for _, feedConfig := range feedConfigs {
		wg.Add(1)
		go func(fc *models.FeedConfig) {
			defer wg.Done()

			articles, err := f.Fetch(ctx, fc)
			resultChan <- result{
				articles: articles,
				err:      err,
				feedName: fc.Name,
			}
		}(feedConfig)
	}

	// すべてのgoroutineが終了するのを待つ
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 結果を集約
	var allArticles []*models.Article
	successCount := 0
	errorCount := 0

	for res := range resultChan {
		if res.err != nil {
			logger.Warn("フィードの取得に失敗",
				"feed_name", res.feedName,
				"error", res.err)
			errorCount++
			continue
		}

		allArticles = append(allArticles, res.articles...)
		successCount++
		logger.Debug("フィードを取得",
			"feed_name", res.feedName,
			"article_count", len(res.articles))
	}

	duration := time.Since(startTime)
	logger.Info("RSSフィードの取得が完了",
		"total_feeds", len(feedConfigs),
		"success", successCount,
		"failed", errorCount,
		"total_articles", len(allArticles),
		"duration_seconds", duration.Seconds())

	return allArticles, nil
}

// Fetch は、単一のフィードから記事を取得する
func (f *Fetcher) Fetch(ctx context.Context, feedConfig *models.FeedConfig) ([]*models.Article, error) {
	// タイムアウト付きコンテキストを作成
	fetchCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	logger.Debug("フィードを取得中",
		"feed_name", feedConfig.Name,
		"feed_url", feedConfig.URL)

	// RSSフィードを取得
	feed, err := f.parser.ParseURLWithContext(feedConfig.URL, fetchCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed %s: %w", feedConfig.Name, err)
	}

	// フィードから記事を抽出
	articles := make([]*models.Article, 0, len(feed.Items))
	for _, item := range feed.Items {
		article := f.convertToArticle(item, feedConfig, feed)
		if article != nil && article.IsValid() {
			articles = append(articles, article)
		}
	}

	return articles, nil
}

// convertToArticle は、gofeed.Itemをmodels.Articleに変換する
func (f *Fetcher) convertToArticle(item *gofeed.Item, feedConfig *models.FeedConfig, feed *gofeed.Feed) *models.Article {
	// IDの決定（GUID > Link）
	id := item.GUID
	if id == "" {
		id = item.Link
	}

	// URLの決定
	url := item.Link
	if url == "" {
		url = item.GUID
	}

	// タイトルの決定
	title := item.Title
	if title == "" {
		title = "（タイトルなし）"
	}

	// 説明文の決定（Description > Content）
	description := item.Description
	if description == "" && item.Content != "" {
		description = item.Content
	}

	// HTMLタグを削除（簡易版）
	description = f.stripHTML(description)

	// 本文
	content := item.Content
	if content == "" {
		content = item.Description
	}
	content = f.stripHTML(content)

	// 著者
	author := ""
	if item.Author != nil {
		author = item.Author.Name
	}

	// 公開日時（Published > Updated > 現在時刻）
	var publishedAt time.Time
	if item.PublishedParsed != nil {
		publishedAt = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		publishedAt = *item.UpdatedParsed
	} else {
		publishedAt = time.Now()
	}

	// 更新日時
	var updatedAt time.Time
	if item.UpdatedParsed != nil {
		updatedAt = *item.UpdatedParsed
	} else {
		updatedAt = publishedAt
	}

	return &models.Article{
		ID:          id,
		Title:       title,
		URL:         url,
		Description: description,
		Content:     content,
		Author:      author,
		PublishedAt: publishedAt,
		UpdatedAt:   updatedAt,
		FeedName:    feedConfig.Name,
		FeedURL:     feedConfig.URL,
		Category:    feedConfig.Category,
		WebhookURL:  feedConfig.WebhookURL, // フィード設定のWebhook URLを引き継ぐ
	}
}

// stripHTML は、HTMLタグを簡易的に削除する
func (f *Fetcher) stripHTML(s string) string {
	// 簡易的なHTMLタグ除去
	// より高度な処理が必要な場合は、goquery等を使用
	s = strings.ReplaceAll(s, "<br>", " ")
	s = strings.ReplaceAll(s, "<br/>", " ")
	s = strings.ReplaceAll(s, "<br />", " ")
	s = strings.ReplaceAll(s, "</p>", " ")

	// タグを削除（簡易版）
	// タグの開始時にスペースを追加して、単語がくっつくのを防ぐ
	inTag := false
	var result strings.Builder
	for _, char := range s {
		if char == '<' {
			inTag = true
			result.WriteRune(' ') // タグの前にスペースを追加
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	// 連続する空白を1つにまとめる
	text := result.String()
	text = strings.Join(strings.Fields(text), " ")

	return strings.TrimSpace(text)
}

// GetFeedInfo は、フィードの情報を取得する（記事は取得しない）
func (f *Fetcher) GetFeedInfo(ctx context.Context, feedURL string) (title string, description string, err error) {
	fetchCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	feed, err := f.parser.ParseURLWithContext(feedURL, fetchCtx)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse feed: %w", err)
	}

	return feed.Title, feed.Description, nil
}

