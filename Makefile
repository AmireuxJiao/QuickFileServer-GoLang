# 彩色输出
G := \033[32m
Y := \033[33m
N := \033[0m

# 基础变量（原有）
BINARY := file-server-cli
SRC    := ./main.go
BUILD  := build

# 新增：测试相关变量（避免硬编码）
DOCS		  := docs
COVER_PROFILE := $(DOCS)/coverage.out  # 覆盖率数据文件
COVER_HTML    := $(DOCS)/coverage.html # 可视化覆盖率报告
TEST_PATH     := ./cmd/...     # 测试文件路径（cmd目录下所有测试）

.PHONY: default
default: help

.PHONY: help
help:                 ## 显示帮助
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
	awk 'BEGIN {FS = ":.*?## "}; {printf " $(G)%-15s$(N) %s\n", $$1, $$2}'

.PHONY: build
build:                ## 编译当前平台
	@mkdir -p $(BUILD)
	@echo "$(G)Building $(BINARY)...$(N)"
	go build -o $(BUILD)/$(BINARY) $(SRC)

.PHONY: clean
clean:                ## 清理构建产物（build目录）
	@echo "$(Y)Cleaning build artifacts...$(N)"
	@rm -rf $(BUILD)

.PHONY: tag
tag:                  ## 打 tag（需 TAG=vx.y.z，如 make tag TAG=v1.0.0）
ifndef TAG
	@echo "$(Y)Usage: make tag TAG=v1.2.3$(N)"
	@exit 1
endif
	git tag -a $(TAG) -m "Release $(TAG)"
	@echo "$(G)Tag $(TAG) created successfully$(N)"

.PHONY: run
run: build	   		  ## 编译并运行
	@echo "$(G)Running $(BINARY)...$(N)"
	./$(BUILD)/$(BINARY)

.PHONY: deps
deps:                 ## 安装/更新项目依赖（含测试依赖）
	@echo "$(G)Updating dependencies...$(N)"
	go mod tidy
	go get github.com/stretchr/testify/assert 2>/dev/null || true

.PHONY: test
test: deps            ## 运行所有测试（cmd目录，含基础覆盖率）
	@echo "$(G)Running all tests (path: $(TEST_PATH))...$(N)"
	go test -v -cover $(TEST_PATH)

.PHONY: test-cover
test-cover: deps      ## 运行测试并生成文本覆盖率报告
	@echo "$(G)Running tests with coverage report...$(N)"
	go test -v -coverprofile=$(COVER_PROFILE) $(TEST_PATH)
	@echo "$(Y)\nCoverage Summary:$(N)"
	go tool cover -func=$(COVER_PROFILE)

.PHONY: test-cover-html
test-cover-html: test-cover  ## 生成可视化HTML覆盖率报告
	mkdir -p $(DOCS)
	@echo "$(G)Generating HTML coverage report...$(N)"
	go tool cover -html=$(COVER_PROFILE) -o=$(COVER_HTML)
	@echo "$(G)HTML report saved to: $(COVER_HTML) (open with browser)$(N)"

.PHONY: clean-test
clean-test:           ## 清理测试产物（覆盖率文件）
	@echo "$(Y)Cleaning test artifacts...$(N)"
	rm -f $(COVER_PROFILE) $(COVER_HTML)
	@echo "$(G)Test artifacts cleaned successfully$(N)"


%: %-help
	@: