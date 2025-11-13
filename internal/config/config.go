package config

import (
	"fmt"
	"os"

	"github.com/ken344/rss-discord-notifier/pkg/models"
	"gopkg.in/yaml.v3"
)

// AppConfig は、アプリケーション全体の設定を管理する構造体
type AppConfig struct {
	// Config は feeds.yaml から読み込まれた設定
	Config *models.Config

	// DiscordWebhookURL は環境変数から読み込まれるDiscord Webhook URL
	DiscordWebhookURL string

	// StateFilePath は状態ファイルのパス
	StateFilePath string

	// LogLevel はログレベル（DEBUG, INFO, WARN, ERROR）
	LogLevel string

	// LogFormat はログフォーマット（json, text）
	LogFormat string
}

// Load は、設定ファイルと環境変数から設定を読み込む
func Load(configFilePath string) (*AppConfig, error) {
	// 1. YAMLファイルを読み込む
	config, err := loadConfigFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// 2. 環境変数を読み込む
	appConfig := &AppConfig{
		Config:            config,
		DiscordWebhookURL: getEnv("DISCORD_WEBHOOK_URL", ""),
		StateFilePath:     getEnv("STATE_FILE_PATH", "./state/state.json"),
		LogLevel:          getEnv("LOG_LEVEL", "INFO"),
		LogFormat:         getEnv("LOG_FORMAT", "json"),
	}

	// 3. バリデーション
	if err := appConfig.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return appConfig, nil
}

// loadConfigFile は、YAMLファイルから設定を読み込む
func loadConfigFile(filePath string) (*models.Config, error) {
	// ファイルを読み込む
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// YAMLをパース
	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// 設定のバリデーション
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 各フィードのWebhook URLの環境変数を展開
	for _, feed := range config.Feeds {
		if feed.WebhookURL != "" {
			feed.WebhookURL = ExpandEnvVars(feed.WebhookURL)
		}
	}

	return &config, nil
}

// Validate は、設定が有効かチェックする
func (a *AppConfig) Validate() error {
	// Discord Webhook URLは必須
	if a.DiscordWebhookURL == "" {
		return fmt.Errorf("DISCORD_WEBHOOK_URL is required")
	}

	// フィードが1つ以上あるかチェック
	if len(a.Config.Feeds) == 0 {
		return fmt.Errorf("at least one feed is required")
	}

	// 有効なフィードが1つ以上あるかチェック
	enabledFeeds := a.Config.GetEnabledFeeds()
	if len(enabledFeeds) == 0 {
		return fmt.Errorf("at least one enabled feed is required")
	}

	// 各フィードの設定をチェック
	for i, feed := range a.Config.Feeds {
		if !feed.IsValid() {
			return fmt.Errorf("feed %d (%s) is invalid", i, feed.Name)
		}
	}

	// ログレベルのバリデーション
	validLogLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
	}
	if !validLogLevels[a.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be DEBUG, INFO, WARN, or ERROR)", a.LogLevel)
	}

	return nil
}

// GetEnabledFeeds は、有効なフィードのリストを返す
func (a *AppConfig) GetEnabledFeeds() []*models.FeedConfig {
	return a.Config.GetEnabledFeeds()
}

// getEnv は、環境変数を取得する。存在しない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

