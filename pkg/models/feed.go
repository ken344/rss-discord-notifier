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
}

// IsValid は、フィード設定が有効かチェックする
func (f *FeedConfig) IsValid() bool {
	// 最低限、名前とURLが必要
	return f.Name != "" && f.URL != ""
}

