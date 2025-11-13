package models

// FeedConfig は、RSSフィードの設定を表すモデル
// feeds.yaml から読み込まれる
type FeedConfig struct {
	// Name はフィードの表示名
	Name string `yaml:"name"`

	// URL はRSSフィードのURL
	URL string `yaml:"url"`

	// Category はフィードのカテゴリ（Tech, News, Blog, Otherなど）
	Category string `yaml:"category"`

	// Enabled はこのフィードが有効かどうか
	Enabled bool `yaml:"enabled"`

	// WebhookURL はこのフィード専用のDiscord Webhook URL（オプション）
	// 環境変数を参照する場合は ${ENV_VAR_NAME} の形式で指定
	// 指定がない場合はデフォルトのWebhook URLが使用される
	WebhookURL string `yaml:"webhook_url,omitempty"`
}

// IsValid は、フィード設定が有効かチェックする
func (f *FeedConfig) IsValid() bool {
	// 最低限、名前とURLが必要
	return f.Name != "" && f.URL != ""
}
