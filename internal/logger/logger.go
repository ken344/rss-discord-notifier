package logger

import (
	"io"
	"log/slog"
	"os"
)

// Logger はアプリケーション全体で使用するロガー
var Logger *slog.Logger

// Level はログレベルを表す型
type Level string

const (
	// LevelDebug はデバッグレベル
	LevelDebug Level = "DEBUG"
	// LevelInfo は情報レベル
	LevelInfo Level = "INFO"
	// LevelWarn は警告レベル
	LevelWarn Level = "WARN"
	// LevelError はエラーレベル
	LevelError Level = "ERROR"
)

// Format はログのフォーマットを表す型
type Format string

const (
	// FormatJSON はJSON形式
	FormatJSON Format = "json"
	// FormatText はテキスト形式
	FormatText Format = "text"
)

// Config はロガーの設定
type Config struct {
	// Level はログレベル
	Level Level

	// Format はログフォーマット
	Format Format

	// Output は出力先（nilの場合は標準出力）
	Output io.Writer
}

// Init は、ロガーを初期化する
func Init(config *Config) {
	if config == nil {
		config = &Config{
			Level:  LevelInfo,
			Format: FormatJSON,
			Output: os.Stdout,
		}
	}

	// デフォルト値の設定
	if config.Output == nil {
		config.Output = os.Stdout
	}

	// ログレベルの変換
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// ハンドラーの作成
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(config.Output, opts)
	case FormatText:
		handler = slog.NewTextHandler(config.Output, opts)
	default:
		handler = slog.NewJSONHandler(config.Output, opts)
	}

	// グローバルロガーを設定
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// Debug はデバッグレベルのログを出力する
func Debug(msg string, args ...any) {
	if Logger != nil {
		Logger.Debug(msg, args...)
	}
}

// Info は情報レベルのログを出力する
func Info(msg string, args ...any) {
	if Logger != nil {
		Logger.Info(msg, args...)
	}
}

// Warn は警告レベルのログを出力する
func Warn(msg string, args ...any) {
	if Logger != nil {
		Logger.Warn(msg, args...)
	}
}

// Error はエラーレベルのログを出力する
func Error(msg string, args ...any) {
	if Logger != nil {
		Logger.Error(msg, args...)
	}
}

// With は、指定された属性を持つ新しいロガーを返す
func With(args ...any) *slog.Logger {
	if Logger != nil {
		return Logger.With(args...)
	}
	return nil
}

// ParseLevel は、文字列をログレベルに変換する
func ParseLevel(s string) Level {
	switch s {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

// ParseFormat は、文字列をフォーマットに変換する
func ParseFormat(s string) Format {
	switch s {
	case "json":
		return FormatJSON
	case "text":
		return FormatText
	default:
		return FormatJSON
	}
}

