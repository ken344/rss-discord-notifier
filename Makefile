# RSS Discord Notifier - Makefile

# 変数定義
BINARY_NAME=notifier
BUILD_DIR=bin
CMD_DIR=cmd/notifier
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# デフォルトターゲット
.DEFAULT_GOAL := help

# カラー出力用
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

# ===== メインターゲット =====

.PHONY: help
help: ## ヘルプを表示
	@echo "$(COLOR_BOLD)RSS Discord Notifier - 利用可能なコマンド$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

.PHONY: all
all: clean lint test build ## すべてのチェックとビルドを実行

.PHONY: build
build: ## バイナリをビルド
	@echo "$(COLOR_BLUE)Building...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "$(COLOR_GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: run
run: ## プログラムを実行
	@echo "$(COLOR_BLUE)Running...$(COLOR_RESET)"
	@$(GO) run $(CMD_DIR)/main.go

.PHONY: test
test: ## テストを実行
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@$(GO) test -v ./...
	@echo "$(COLOR_GREEN)✓ Tests passed$(COLOR_RESET)"

.PHONY: test-coverage
test-coverage: ## カバレッジ付きでテストを実行
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	@$(GO) test -cover -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: coverage.html$(COLOR_RESET)"

.PHONY: lint
lint: ## Linterを実行
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Lint passed$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Run: make install-tools$(COLOR_RESET)"; \
	fi

.PHONY: fmt
fmt: ## コードをフォーマット
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@$(GO) fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## go vet を実行
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@$(GO) vet ./...
	@echo "$(COLOR_GREEN)✓ Vet passed$(COLOR_RESET)"

.PHONY: tidy
tidy: ## go mod tidy を実行
	@echo "$(COLOR_BLUE)Tidying modules...$(COLOR_RESET)"
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ Modules tidied$(COLOR_RESET)"

.PHONY: clean
clean: ## ビルド成果物を削除
	@echo "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(COLOR_GREEN)✓ Cleaned$(COLOR_RESET)"

# ===== セットアップ =====

.PHONY: setup
setup: ## 初期セットアップ
	@echo "$(COLOR_BLUE)Setting up project...$(COLOR_RESET)"
	@$(GO) mod download
	@mkdir -p $(BUILD_DIR) state configs
	@if [ ! -f configs/feeds.yaml ]; then \
		cp configs/feeds.example.yaml configs/feeds.yaml; \
		echo "$(COLOR_GREEN)✓ Created configs/feeds.yaml$(COLOR_RESET)"; \
	fi
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "$(COLOR_GREEN)✓ Created .env$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)⚠ Please edit .env and set your DISCORD_WEBHOOK_URL$(COLOR_RESET)"; \
	fi
	@echo "$(COLOR_GREEN)✓ Setup complete$(COLOR_RESET)"

.PHONY: install-tools
install-tools: ## 開発ツールをインストール
	@echo "$(COLOR_BLUE)Installing development tools...$(COLOR_RESET)"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(COLOR_GREEN)✓ Tools installed$(COLOR_RESET)"

# ===== Docker（オプション） =====

.PHONY: docker-build
docker-build: ## Dockerイメージをビルド
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	@docker build -t rss-discord-notifier:latest .
	@echo "$(COLOR_GREEN)✓ Docker image built$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## Dockerコンテナを実行
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	@docker run --rm --env-file .env rss-discord-notifier:latest

# ===== クロスコンパイル =====

.PHONY: build-all
build-all: build-linux build-darwin build-windows ## 全プラットフォーム向けにビルド

.PHONY: build-linux
build-linux: ## Linux用にビルド
	@echo "$(COLOR_BLUE)Building for Linux...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/main.go
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64$(COLOR_RESET)"

.PHONY: build-darwin
build-darwin: ## macOS用にビルド
	@echo "$(COLOR_BLUE)Building for macOS...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/main.go
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-*$(COLOR_RESET)"

.PHONY: build-windows
build-windows: ## Windows用にビルド
	@echo "$(COLOR_BLUE)Building for Windows...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/main.go
	@echo "$(COLOR_GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe$(COLOR_RESET)"

# ===== ユーティリティ =====

.PHONY: deps
deps: ## 依存関係を表示
	@$(GO) list -m all

.PHONY: deps-update
deps-update: ## 依存関係を最新化
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)⚠ Please run 'make test' to verify$(COLOR_RESET)"

.PHONY: deps-check
deps-check: ## 依存関係の更新をチェック
	@echo "$(COLOR_BLUE)Checking for dependency updates...$(COLOR_RESET)"
	@$(GO) list -u -m all | grep '\['

.PHONY: deps-graph
deps-graph: ## 依存関係のグラフを表示
	@echo "$(COLOR_BLUE)Dependency graph:$(COLOR_RESET)"
	@$(GO) mod graph

.PHONY: check
check: fmt vet lint test ## すべてのチェックを実行（ビルドなし）

.PHONY: init-state
init-state: ## 状態ファイルを初期化（削除）
	@echo "$(COLOR_YELLOW)⚠ This will delete the state file$(COLOR_RESET)"
	@rm -f state/state.json
	@echo "$(COLOR_GREEN)✓ State file deleted$(COLOR_RESET)"

.PHONY: version
version: ## Goバージョンを表示
	@$(GO) version

