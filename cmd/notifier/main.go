package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/ken344/rss-discord-notifier/internal/config"
	"github.com/ken344/rss-discord-notifier/internal/discord"
	"github.com/ken344/rss-discord-notifier/internal/feed"
	"github.com/ken344/rss-discord-notifier/internal/logger"
	"github.com/ken344/rss-discord-notifier/internal/state"
	"github.com/ken344/rss-discord-notifier/pkg/models"
)

const (
	// 初回実行時に通知する記事数の上限
	maxArticlesOnFirstRun = 5
)

func main() {
	// エラーが発生した場合は、終了コード1で終了
	if err := run(); err != nil {
		logger.Error("アプリケーションの実行に失敗しました", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// 実行開始時刻を記録
	startTime := time.Now()

	// コンテキストを作成
	ctx := context.Background()

	// 1. 設定を読み込む
	logger.Info("設定を読み込んでいます...")
	configFilePath := getEnv("CONFIG_FILE_PATH", "./configs/feeds.yaml")
	appConfig, err := config.Load(configFilePath)
	if err != nil {
		return fmt.Errorf("設定の読み込みに失敗: %w", err)
	}

	// 2. ロガーを初期化
	logger.Init(&logger.Config{
		Level:  logger.ParseLevel(appConfig.LogLevel),
		Format: logger.ParseFormat(appConfig.LogFormat),
	})

	logger.Info("RSS Discord Notifierを起動しました",
		"version", "1.0.0",
		"feeds_count", len(appConfig.GetEnabledFeeds()))

	// 3. 状態管理マネージャーを初期化
	logger.Info("状態を読み込んでいます...")
	stateManager := state.NewManager(appConfig.StateFilePath)
	if err := stateManager.Load(); err != nil {
		return fmt.Errorf("状態の読み込みに失敗: %w", err)
	}

	// 初回実行かチェック
	isFirstRun := stateManager.IsFirstRun()
	if isFirstRun {
		logger.Info("初回実行を検出しました（最新記事のみを通知します）")
	}

	// 4. RSSフィード取得器を初期化
	timeout := time.Duration(appConfig.Config.Notification.TimeoutSeconds) * time.Second
	fetcher := feed.NewFetcher(timeout)

	// 5. フィードから記事を取得
	logger.Info("RSSフィードから記事を取得しています...")
	enabledFeeds := appConfig.GetEnabledFeeds()
	allArticles, err := fetcher.FetchAll(ctx, enabledFeeds)
	if err != nil {
		return fmt.Errorf("フィードの取得に失敗: %w", err)
	}

	logger.Info("記事の取得が完了しました", "total_articles", len(allArticles))

	// 6. 新規記事をフィルタリング
	newArticles := filterNewArticles(allArticles, stateManager)
	logger.Info("新規記事を検出しました", "new_articles", len(newArticles))

	// 初回実行の場合は、最新N件のみに制限
	if isFirstRun && len(newArticles) > maxArticlesOnFirstRun {
		logger.Info("初回実行のため、記事数を制限します",
			"before", len(newArticles),
			"after", maxArticlesOnFirstRun)
		newArticles = limitArticles(newArticles, maxArticlesOnFirstRun)
	}

	// 通知する記事数の上限をチェック
	maxArticles := appConfig.Config.Notification.MaxArticlesPerRun
	if len(newArticles) > maxArticles {
		logger.Warn("記事数が上限を超えています。古い記事から制限します。",
			"count", len(newArticles),
			"max", maxArticles)
		newArticles = limitArticles(newArticles, maxArticles)
	}

	// 7. Discordに通知
	if len(newArticles) > 0 {
		logger.Info("Discordに通知を送信しています...", "count", len(newArticles))

		// レート制限設定
		rateLimit := time.Duration(appConfig.Config.Notification.RateLimitMs) * time.Millisecond

		// 記事を通知（古い順に）
		sortedArticles := sortArticlesByPublishedAt(newArticles)

		// 記事ごとに適切なWebhook URLで通知
		successCount := 0
		for i, article := range sortedArticles {
			// Webhook URLの決定（記事設定 > デフォルト）
			webhookURL := article.WebhookURL
			if webhookURL == "" {
				webhookURL = appConfig.DiscordWebhookURL
			}

			// Notifierを作成（Webhook URLごとに作成）
			notifier := discord.NewNotifier(webhookURL, rateLimit)

			// 記事を送信
			if err := notifier.SendArticle(ctx, article); err != nil {
				logger.Error("記事の通知に失敗",
					"title", article.Title,
					"feed", article.FeedName,
					"error", err)
				continue
			}

			// 8. 通知済み記事を状態に記録
			stateManager.MarkAsNotified(article)
			successCount++

			// レート制限対策（最後の記事以外）
			if i < len(sortedArticles)-1 {
				select {
				case <-time.After(rateLimit):
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		logger.Info("通知が完了しました",
			"total", len(sortedArticles),
			"success", successCount,
			"failed", len(sortedArticles)-successCount)
	} else {
		logger.Info("通知する新規記事がありません")
	}

	// 9. 統計情報を更新
	duration := time.Since(startTime)
	stateManager.UpdateStatistics(len(enabledFeeds), duration.Seconds())

	// 10. 状態を保存
	logger.Info("状態を保存しています...")
	if err := stateManager.Save(); err != nil {
		return fmt.Errorf("状態の保存に失敗: %w", err)
	}

	// サマリーを表示
	logger.Info("実行サマリー",
		"feeds_checked", len(enabledFeeds),
		"total_articles", len(allArticles),
		"new_articles", len(newArticles),
		"notified", len(newArticles),
		"duration_seconds", duration.Seconds())

	return nil
}

// filterNewArticles は、新規記事のみをフィルタリングする
func filterNewArticles(articles []*models.Article, stateManager *state.Manager) []*models.Article {
	newArticles := make([]*models.Article, 0)

	for _, article := range articles {
		// 記事の検証
		if !article.IsValid() {
			logger.Warn("無効な記事をスキップ",
				"title", article.Title,
				"id", article.ID)
			continue
		}

		// 既読チェック
		if stateManager.IsArticleNotified(article.FeedURL, article.ID) {
			logger.Debug("既読記事をスキップ",
				"title", article.Title,
				"feed", article.FeedName)
			continue
		}

		newArticles = append(newArticles, article)
	}

	return newArticles
}

// limitArticles は、記事を最新N件に制限する
func limitArticles(articles []*models.Article, limit int) []*models.Article {
	if len(articles) <= limit {
		return articles
	}

	// 公開日時でソート（新しい順）
	sorted := make([]*models.Article, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PublishedAt.After(sorted[j].PublishedAt)
	})

	// 最新N件を返す
	return sorted[:limit]
}

// sortArticlesByPublishedAt は、記事を公開日時でソートする（古い順）
func sortArticlesByPublishedAt(articles []*models.Article) []*models.Article {
	sorted := make([]*models.Article, len(articles))
	copy(sorted, articles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PublishedAt.Before(sorted[j].PublishedAt)
	})
	return sorted
}

// getEnv は、環境変数を取得する。存在しない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
