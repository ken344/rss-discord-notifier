package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ken344/rss-discord-notifier/internal/logger"
	"github.com/ken344/rss-discord-notifier/pkg/models"
)

// Manager は、状態管理を担当する構造体
type Manager struct {
	// filePath は状態ファイルのパス
	filePath string

	// state は現在の状態
	state *models.State

	// maxArticlesPerFeed はフィードごとに保持する最大記事数
	maxArticlesPerFeed int

	// cleanupDays は古い記事を削除する日数
	cleanupDays int
}

// NewManager は、新しい状態管理マネージャーを作成する
func NewManager(filePath string) *Manager {
	return &Manager{
		filePath:           filePath,
		state:              models.NewState(),
		maxArticlesPerFeed: 100, // フィードごとに最新100件を保持
		cleanupDays:        30,  // 30日より古い記事は削除
	}
}

// Load は、状態ファイルから状態を読み込む
func (m *Manager) Load() error {
	// ファイルが存在するかチェック
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		logger.Info("状態ファイルが存在しないため、新規作成します", "path", m.filePath)
		m.state = models.NewState()
		return nil
	}

	// ファイルを読み込む
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// JSONをパース
	var state models.State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	m.state = &state
	logger.Info("状態ファイルを読み込みました",
		"path", m.filePath,
		"last_update", m.state.LastUpdate,
		"feeds_count", len(m.state.Feeds))

	return nil
}

// Save は、現在の状態をファイルに保存する
func (m *Manager) Save() error {
	// 保存前のクリーンアップ
	m.cleanup()

	// 最終更新日時を更新
	m.state.LastUpdate = time.Now()

	// JSONにエンコード（インデント付き）
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// ディレクトリが存在しない場合は作成
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// ファイルに書き込む
	if err := os.WriteFile(m.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	logger.Info("状態ファイルを保存しました",
		"path", m.filePath,
		"feeds_count", len(m.state.Feeds),
		"total_articles", m.state.Statistics.TotalArticlesNotified)

	return nil
}

// IsArticleNotified は、指定された記事が通知済みかチェックする
func (m *Manager) IsArticleNotified(feedURL, articleID string) bool {
	// フィードが存在しない場合はfalseを返す（新規作成しない）
	feedState, exists := m.state.Feeds[feedURL]
	if !exists {
		return false
	}
	return feedState.IsArticleNotified(articleID)
}

// MarkAsNotified は、記事を通知済みとしてマークする
func (m *Manager) MarkAsNotified(article *models.Article) {
	feedState := m.state.GetFeedState(article.FeedURL)
	feedState.AddNotifiedArticle(article)

	// 統計情報を更新
	m.state.Statistics.TotalArticlesNotified++
}

// GetFeedState は、指定されたフィードの状態を取得する
func (m *Manager) GetFeedState(feedURL string) *models.FeedState {
	return m.state.GetFeedState(feedURL)
}

// UpdateStatistics は、統計情報を更新する
func (m *Manager) UpdateStatistics(feedsChecked int, duration float64) {
	m.state.Statistics.TotalFeedsChecked += feedsChecked
	m.state.Statistics.LastRunDuration = duration
}

// GetState は、現在の状態を返す
func (m *Manager) GetState() *models.State {
	return m.state
}

// cleanup は、古い記事情報を削除する
func (m *Manager) cleanup() {
	if m.cleanupDays <= 0 {
		return
	}

	logger.Debug("状態のクリーンアップを実行",
		"cleanup_days", m.cleanupDays,
		"max_articles_per_feed", m.maxArticlesPerFeed)

	for feedURL, feedState := range m.state.Feeds {
		// 古い記事を削除
		beforeCount := len(feedState.NotifiedArticles)
		feedState.CleanupOldArticles(m.cleanupDays)

		// 記事数を制限
		feedState.LimitArticleCount(m.maxArticlesPerFeed)
		afterCount := len(feedState.NotifiedArticles)

		if beforeCount != afterCount {
			logger.Debug("フィードの記事をクリーンアップ",
				"feed_url", feedURL,
				"before", beforeCount,
				"after", afterCount)
		}
	}
}

// Reset は、状態をリセットする（主にテスト用）
func (m *Manager) Reset() {
	m.state = models.NewState()
	logger.Info("状態をリセットしました")
}

// GetNotifiedArticleCount は、フィードごとの通知済み記事数を返す
func (m *Manager) GetNotifiedArticleCount(feedURL string) int {
	feedState, exists := m.state.Feeds[feedURL]
	if !exists {
		return 0
	}
	return len(feedState.NotifiedArticles)
}

// IsFirstRun は、初回実行かどうかを判定する
func (m *Manager) IsFirstRun() bool {
	// 状態ファイルが存在せず、フィードが1つもない場合は初回実行
	return len(m.state.Feeds) == 0
}

// SetMaxArticlesPerFeed は、フィードごとの最大記事数を設定する
func (m *Manager) SetMaxArticlesPerFeed(max int) {
	if max > 0 {
		m.maxArticlesPerFeed = max
	}
}

// SetCleanupDays は、クリーンアップ対象の日数を設定する
func (m *Manager) SetCleanupDays(days int) {
	if days > 0 {
		m.cleanupDays = days
	}
}
