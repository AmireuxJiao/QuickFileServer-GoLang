# 彩色输出
G := \033[32m
Y := \033[33m
N := \033[0m

BINARY := quick_server
SRC    := ./cmd/main.go
BUILD  := build

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
clean:                ## 清理构建产物
	@echo "$(Y)Cleaning...$(N)"
	@rm -rf $(BUILD)

.PHONY: tag
tag:                  ## 打 tag（需 TAG=vx.y.z）
ifndef TAG
	@echo "Usage: make tag TAG=v1.2.3"
	@exit 1
endif
	git tag -a $(TAG) -m "Release $(TAG)"

%: %-help
	@: