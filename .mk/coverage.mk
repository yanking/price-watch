COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

## coverage: 生成覆盖率报告
coverage:
	@echo "Running coverage..."
	@$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -func=$(COVERAGE_FILE)

## coverage-html: 生成 HTML 覆盖率报告
coverage-html: coverage
	@echo "Generating coverage HTML..."
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"
