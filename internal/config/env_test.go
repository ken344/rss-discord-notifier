package config

import (
	"os"
	"testing"
)

// TestExpandEnvVars は、環境変数展開をテストする
func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		want     string
	}{
		{
			name:    "環境変数なし",
			input:   "plain text",
			envVars: map[string]string{},
			want:    "plain text",
		},
		{
			name:  "単一の環境変数",
			input: "${TEST_VAR}",
			envVars: map[string]string{
				"TEST_VAR": "test_value",
			},
			want: "test_value",
		},
		{
			name:  "複数の環境変数",
			input: "${VAR1}/${VAR2}",
			envVars: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			want: "value1/value2",
		},
		{
			name:  "環境変数と通常のテキスト",
			input: "https://example.com/${API_KEY}/resource",
			envVars: map[string]string{
				"API_KEY": "secret123",
			},
			want: "https://example.com/secret123/resource",
		},
		{
			name:    "存在しない環境変数",
			input:   "${NONEXISTENT}",
			envVars: map[string]string{},
			want:    "${NONEXISTENT}", // 元の文字列のまま
		},
		{
			name:  "Discord Webhook URL",
			input: "${DISCORD_WEBHOOK_URL_TECH}",
			envVars: map[string]string{
				"DISCORD_WEBHOOK_URL_TECH": "https://discord.com/api/webhooks/123/abc",
			},
			want: "https://discord.com/api/webhooks/123/abc",
		},
		{
			name:    "空文字列",
			input:   "",
			envVars: map[string]string{},
			want:    "",
		},
		{
			name:  "一部の環境変数が存在しない",
			input: "${EXISTS}/${NOT_EXISTS}",
			envVars: map[string]string{
				"EXISTS": "value",
			},
			want: "value/${NOT_EXISTS}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テスト用の環境変数を設定
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			got := ExpandEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("ExpandEnvVars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestExpandEnvVarsPattern は、環境変数パターンマッチングをテストする
func TestExpandEnvVarsPattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "大文字とアンダースコア",
			input: "${TEST_VAR_123}",
			want:  "${TEST_VAR_123}", // 環境変数が存在しないので変更なし
		},
		{
			name:  "小文字（マッチしない）",
			input: "${test_var}",
			want:  "${test_var}", // パターンにマッチしないので変更なし
		},
		{
			name:  "数字から始まる（マッチしない）",
			input: "${123_VAR}",
			want:  "${123_VAR}", // パターンにマッチしないので変更なし
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandEnvVars(tt.input)
			if got != tt.want {
				t.Errorf("ExpandEnvVars(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

