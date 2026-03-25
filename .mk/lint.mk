# golangci-lint 工具变量
GOLANGCI_LINT := golangci-lint

## lint: 运行代码检查
lint:
	@echo "Running linter..."
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run ./...; \
	else \
		echo "$(GOLANGCI_LINT) not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## lint-fix: 运行代码检查并自动修复
lint-fix:
	@echo "Running linter with auto-fix..."
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run --fix ./...; \
	else \
		echo "$(GOLANGCI_LINT) not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
