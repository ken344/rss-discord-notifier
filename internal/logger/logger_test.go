package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestInit は、ロガーの初期化をテストする
func TestInit(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "デフォルト設定",
			config: nil,
		},
		{
			name: "JSON形式、INFOレベル",
			config: &Config{
				Level:  LevelInfo,
				Format: FormatJSON,
			},
		},
		{
			name: "テキスト形式、DEBUGレベル",
			config: &Config{
				Level:  LevelDebug,
				Format: FormatText,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// パニックが起きないことを確認
			Init(tt.config)
			
			if Logger == nil {
				t.Error("Logger should not be nil after Init()")
			}
		})
	}
}

// TestLogLevels は、各ログレベルでの出力をテストする
func TestLogLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		logFunc  func()
		contains string
	}{
		{
			name:     "INFOレベル",
			level:    LevelInfo,
			logFunc:  func() { Info("test info message") },
			contains: "test info message",
		},
		{
			name:     "WARNレベル",
			level:    LevelWarn,
			logFunc:  func() { Warn("test warn message") },
			contains: "test warn message",
		},
		{
			name:     "ERRORレベル",
			level:    LevelError,
			logFunc:  func() { Error("test error message") },
			contains: "test error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// バッファに出力を書き込む
			buf := &bytes.Buffer{}
			Init(&Config{
				Level:  tt.level,
				Format: FormatJSON,
				Output: buf,
			})

			// ログを出力
			tt.logFunc()

			// 出力にメッセージが含まれているか確認
			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("log output should contain %q, got %q", tt.contains, output)
			}

			// JSON形式として正しいか確認
			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Errorf("log output should be valid JSON: %v", err)
			}
		})
	}
}

// TestLogWithAttributes は、属性付きログをテストする
func TestLogWithAttributes(t *testing.T) {
	buf := &bytes.Buffer{}
	Init(&Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: buf,
	})

	// 属性付きでログを出力
	Info("test message", "key1", "value1", "key2", 123)

	output := buf.String()

	// メッセージが含まれているか
	if !strings.Contains(output, "test message") {
		t.Errorf("output should contain 'test message', got %q", output)
	}

	// JSON形式として正しいか
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("output should be valid JSON: %v", err)
	}

	// 属性が含まれているか
	if logEntry["key1"] != "value1" {
		t.Errorf("log should contain key1=value1, got %v", logEntry["key1"])
	}
}

// TestParseLevel は、ログレベルのパースをテストする
func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  Level
	}{
		{"DEBUG", LevelDebug},
		{"INFO", LevelInfo},
		{"WARN", LevelWarn},
		{"ERROR", LevelError},
		{"INVALID", LevelInfo}, // デフォルト
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.want {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseFormat は、フォーマットのパースをテストする
func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
	}{
		{"json", FormatJSON},
		{"text", FormatText},
		{"invalid", FormatJSON}, // デフォルト
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseFormat(tt.input)
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestDebugLevel は、DEBUGレベルが正しくフィルタリングされるかテストする
func TestDebugLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	
	// INFOレベルで初期化（DEBUGは出力されないはず）
	Init(&Config{
		Level:  LevelInfo,
		Format: FormatJSON,
		Output: buf,
	})

	Debug("this should not appear")
	Info("this should appear")

	output := buf.String()

	// DEBUGメッセージは出力されていないはず
	if strings.Contains(output, "this should not appear") {
		t.Error("DEBUG message should not appear when level is INFO")
	}

	// INFOメッセージは出力されているはず
	if !strings.Contains(output, "this should appear") {
		t.Error("INFO message should appear when level is INFO")
	}
}

