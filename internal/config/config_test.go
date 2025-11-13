package config

import (
	"os"
	"testing"

	"github.com/ken344/rss-discord-notifier/pkg/models"
)

// TestGetEnv は、環境変数取得のテスト
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "環境変数が設定されている場合",
			key:          "TEST_KEY",
			defaultValue: "default",
			envValue:     "test_value",
			want:         "test_value",
		},
		{
			name:         "環境変数が設定されていない場合",
			key:          "TEST_KEY_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の環境変数を設定
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := getEnv(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLoadConfigFile は、設定ファイル読み込みのテスト
func TestLoadConfigFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "存在しないファイル",
			path:    "nonexistent.yaml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadConfigFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfigFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestAppConfig_Validate は、設定のバリデーションテスト
func TestAppConfig_Validate(t *testing.T) {
	// テスト用の最小限のConfig
	minimalConfig := &AppConfig{
		DiscordWebhookURL: "https://discord.com/api/webhooks/test",
		LogLevel:          "INFO",
		Config: &models.Config{
			Version: "1.0",
			Feeds: []*models.FeedConfig{
				{
					Name:    "Test Feed",
					URL:     "https://example.com/feed",
					Enabled: true,
				},
			},
		},
	}

	tests := []struct {
		name    string
		config  *AppConfig
		wantErr bool
	}{
		{
			name: "Discord Webhook URLが空",
			config: &AppConfig{
				DiscordWebhookURL: "",
				LogLevel:          "INFO",
				Config: &models.Config{
					Version: "1.0",
					Feeds: []*models.FeedConfig{
						{
							Name:    "Test Feed",
							URL:     "https://example.com/feed",
							Enabled: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "無効なログレベル",
			config: &AppConfig{
				DiscordWebhookURL: "https://discord.com/api/webhooks/test",
				LogLevel:          "INVALID",
				Config: &models.Config{
					Version: "1.0",
					Feeds: []*models.FeedConfig{
						{
							Name:    "Test Feed",
							URL:     "https://example.com/feed",
							Enabled: true,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "正常な設定",
			config:  minimalConfig,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AppConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

