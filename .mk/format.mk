## fmt: 格式化代码并显示修改的文件
fmt:
	@echo "Formatting code..."
	@gofmt -w -l . | head -10 || true

## vet: 运行静态分析（建议先执行 fmt 避免无效警告）
vet:
	@echo "Running go vet..."
	@go vet ./...
