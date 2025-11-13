package models

// Config は、アプリケーション全体の設定を表すモデル
type Config struct {
	// Version は設定ファイルのバージョン
	Version string `yaml:"version"`

	// Notification は通知に関する設定
	Notification *NotificationConfig `yaml:"notification"`

	// Feeds は監視するRSSフィードのリスト
	Feeds []*FeedConfig `yaml:"feeds"`
}

// NotificationConfig は、通知に関する設定を表すモデル
type NotificationConfig struct {
	// MaxArticlesPerRun は1回の実行で通知する最大記事数
	MaxArticlesPerRun int `yaml:"max_articles_per_run"`

	// TimeoutSeconds はフィード取得のタイムアウト（秒）
	TimeoutSeconds int `yaml:"timeout_seconds"`

	// RateLimitMs はDiscord通知間隔（ミリ秒）
	RateLimitMs int `yaml:"rate_limit_ms"`
}

// GetEnabledFeeds は、有効なフィードのみを返す
func (c *Config) GetEnabledFeeds() []*FeedConfig {
	enabledFeeds := make([]*FeedConfig, 0)
	
	for _, feed := range c.Feeds {
		if feed.Enabled {
			enabledFeeds = append(enabledFeeds, feed)
		}
	}
	
	return enabledFeeds
}

// Validate は、設定が有効かチェックする
func (c *Config) Validate() error {
	if c.Notification == nil {
		// デフォルト値を設定
		c.Notification = &NotificationConfig{
			MaxArticlesPerRun: 10,
			TimeoutSeconds:    30,
			RateLimitMs:       1000,
		}
	}

	// 設定値の妥当性チェック
	if c.Notification.MaxArticlesPerRun <= 0 {
		c.Notification.MaxArticlesPerRun = 10
	}

	if c.Notification.TimeoutSeconds <= 0 {
		c.Notification.TimeoutSeconds = 30
	}

	if c.Notification.RateLimitMs < 0 {
		c.Notification.RateLimitMs = 1000
	}

	return nil
}

