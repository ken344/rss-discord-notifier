package config

import (
	"os"
	"regexp"
	"strings"
)

// envVarPattern は環境変数参照のパターン ${ENV_VAR}
var envVarPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)

// ExpandEnvVars は、文字列内の環境変数参照を展開する
// ${ENV_VAR_NAME} の形式で記述された環境変数を実際の値に置き換える
//
// 例:
//   - "${DISCORD_WEBHOOK_URL}" → 環境変数の値
//   - "https://example.com/${API_KEY}" → "https://example.com/actual_key"
//   - "plain text" → "plain text" (変更なし)
func ExpandEnvVars(s string) string {
	if s == "" {
		return s
	}

	// ${ENV_VAR} パターンを検索して置換
	result := envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// ${ENV_VAR} から ENV_VAR を抽出
		envName := strings.TrimPrefix(match, "${")
		envName = strings.TrimSuffix(envName, "}")

		// 環境変数の値を取得
		if value := os.Getenv(envName); value != "" {
			return value
		}

		// 環境変数が存在しない場合は元の文字列をそのまま返す
		return match
	})

	return result
}
