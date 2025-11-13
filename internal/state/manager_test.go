package state

import (
	"fmt"
	"os"
	"path/filepath"
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

// TestNewManager は、マネージャーの作成をテストする
func TestNewManager(t *testing.T) {
	manager := NewManager("test.json")

	if manager == nil {
		t.Fatal("NewManager() should not return nil")
	}

	if manager.filePath != "test.json" {
		t.Errorf("filePath = %v, want %v", manager.filePath, "test.json")
	}

	if manager.state == nil {
		t.Error("state should not be nil")
	}
}

// TestLoadAndSave は、状態の読み書きをテストする
func TestLoadAndSave(t *testing.T) {
	// 一時ディレクトリを作成
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "state.json")

	// マネージャーを作成
	manager := NewManager(stateFile)

	// 初回読み込み（ファイルが存在しない）
	if err := manager.Load(); err != nil {
		t.Fatalf("Load() should succeed even if file does not exist: %v", err)
	}

	// 記事を追加
	article := &models.Article{
		ID:          "article-1",
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		PublishedAt: time.Now(),
		FeedURL:     "https://example.com/feed",
	}
	manager.MarkAsNotified(article)

	// 保存
	if err := manager.Save(); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// ファイルが作成されたか確認
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Fatal("state file should exist after Save()")
	}

	// 新しいマネージャーで読み込み
	manager2 := NewManager(stateFile)
	if err := manager2.Load(); err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// 記事が存在するか確認
	if !manager2.IsArticleNotified("https://example.com/feed", "article-1") {
		t.Error("article should be marked as notified after loading")
	}
}

// TestIsArticleNotified は、既読チェックをテストする
func TestIsArticleNotified(t *testing.T) {
	manager := NewManager("test.json")

	feedURL := "https://example.com/feed"
	articleID := "article-1"

	// 初期状態では未読
	if manager.IsArticleNotified(feedURL, articleID) {
		t.Error("article should not be notified initially")
	}

	// 記事をマーク
	article := &models.Article{
		ID:          articleID,
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		PublishedAt: time.Now(),
		FeedURL:     feedURL,
	}
	manager.MarkAsNotified(article)

	// マーク後は既読
	if !manager.IsArticleNotified(feedURL, articleID) {
		t.Error("article should be notified after marking")
	}
}

// TestMarkAsNotified は、通知済みマーキングをテストする
func TestMarkAsNotified(t *testing.T) {
	manager := NewManager("test.json")

	article := &models.Article{
		ID:          "article-1",
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		PublishedAt: time.Now(),
		FeedURL:     "https://example.com/feed",
	}

	// 初期の統計
	initialCount := manager.state.Statistics.TotalArticlesNotified

	// マーク
	manager.MarkAsNotified(article)

	// 統計が更新されたか
	if manager.state.Statistics.TotalArticlesNotified != initialCount+1 {
		t.Errorf("TotalArticlesNotified = %d, want %d",
			manager.state.Statistics.TotalArticlesNotified, initialCount+1)
	}

	// 既読チェック
	if !manager.IsArticleNotified("https://example.com/feed", "article-1") {
		t.Error("article should be marked as notified")
	}
}

// TestCleanup は、クリーンアップをテストする
func TestCleanup(t *testing.T) {
	manager := NewManager("test.json")
	manager.SetCleanupDays(7)
	manager.SetMaxArticlesPerFeed(5)

	feedURL := "https://example.com/feed"

	// 古い記事を追加（8日前）
	oldArticle := &models.Article{
		ID:          "old-article",
		Title:       "Old Article",
		URL:         "https://example.com/old",
		PublishedAt: time.Now().AddDate(0, 0, -8),
		FeedURL:     feedURL,
	}
	manager.MarkAsNotified(oldArticle)

	// 新しい記事を追加
	for i := 0; i < 10; i++ {
		article := &models.Article{
			ID:          fmt.Sprintf("article-%d", i),
			Title:       fmt.Sprintf("Article %d", i),
			URL:         fmt.Sprintf("https://example.com/article-%d", i),
			PublishedAt: time.Now(),
			FeedURL:     feedURL,
		}
		manager.MarkAsNotified(article)
	}

	// クリーンアップ前の記事数
	beforeCount := manager.GetNotifiedArticleCount(feedURL)
	if beforeCount != 11 { // 古い記事1 + 新しい記事10
		t.Errorf("before cleanup: count = %d, want 11", beforeCount)
	}

	// クリーンアップ実行（Saveで自動実行される）
	manager.cleanup()

	// クリーンアップ後の記事数（最大5件）
	afterCount := manager.GetNotifiedArticleCount(feedURL)
	if afterCount > 5 {
		t.Errorf("after cleanup: count = %d, want <= 5", afterCount)
	}

	// 古い記事が削除されたか
	if manager.IsArticleNotified(feedURL, "old-article") {
		t.Error("old article should be removed after cleanup")
	}
}

// TestIsFirstRun は、初回実行判定をテストする
func TestIsFirstRun(t *testing.T) {
	manager := NewManager("test.json")

	// 初期状態では初回実行
	if !manager.IsFirstRun() {
		t.Error("should be first run initially")
	}

	// 記事を追加
	article := &models.Article{
		ID:          "article-1",
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		PublishedAt: time.Now(),
		FeedURL:     "https://example.com/feed",
	}
	manager.MarkAsNotified(article)

	// 記事追加後は初回実行ではない
	if manager.IsFirstRun() {
		t.Error("should not be first run after adding article")
	}
}

// TestUpdateStatistics は、統計情報の更新をテストする
func TestUpdateStatistics(t *testing.T) {
	manager := NewManager("test.json")

	// 初期値
	if manager.state.Statistics.TotalFeedsChecked != 0 {
		t.Error("TotalFeedsChecked should be 0 initially")
	}

	// 統計更新
	manager.UpdateStatistics(5, 1.23)

	if manager.state.Statistics.TotalFeedsChecked != 5 {
		t.Errorf("TotalFeedsChecked = %d, want 5", manager.state.Statistics.TotalFeedsChecked)
	}

	if manager.state.Statistics.LastRunDuration != 1.23 {
		t.Errorf("LastRunDuration = %f, want 1.23", manager.state.Statistics.LastRunDuration)
	}
}

// TestReset は、リセット機能をテストする
func TestReset(t *testing.T) {
	manager := NewManager("test.json")

	// 記事を追加
	article := &models.Article{
		ID:          "article-1",
		Title:       "Test Article",
		URL:         "https://example.com/article-1",
		PublishedAt: time.Now(),
		FeedURL:     "https://example.com/feed",
	}
	manager.MarkAsNotified(article)

	// 記事が存在することを確認
	if !manager.IsArticleNotified("https://example.com/feed", "article-1") {
		t.Error("article should exist before reset")
	}

	// リセット
	manager.Reset()

	// 記事が削除されたことを確認
	if manager.IsArticleNotified("https://example.com/feed", "article-1") {
		t.Error("article should not exist after reset")
	}

	// Feedsマップの状態を確認
	feedsCount := len(manager.state.Feeds)
	t.Logf("After reset: Feeds count = %d", feedsCount)

	// 初回実行状態に戻っているか
	if !manager.IsFirstRun() {
		t.Errorf("should be first run after reset (feeds count: %d)", feedsCount)
	}
}

