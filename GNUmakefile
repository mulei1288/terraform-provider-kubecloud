default: fmt lint install generate

build:
	GOSUMDB=sum.golang.org go build -v ./...

install: build
	go install -v ./...

install-local: build
	@echo "正在构建并安装 provider 到本地 Terraform 插件目录..."
	@mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bingocloud/0.1.0/darwin_arm64
	GOSUMDB=sum.golang.org go build -v -o terraform-provider-bingocloud .
	cp terraform-provider-bingocloud ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bingocloud/0.1.0/darwin_arm64/terraform-provider-bingocloud_v0.1.0
	@echo "✓ Provider 已成功安装到: ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bingocloud/0.1.0/darwin_arm64/"
	@ls -lh ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bingocloud/0.1.0/darwin_arm64/

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout 120m ./...

.PHONY: fmt lint test testacc build install install-local generate


# 自动获取当前分支
CURRENT_BRANCH := $(shell git branch --show-current 2>/dev/null || echo "unknown")
COMMIT_TIME := $(shell date '+%Y-%m-%d_%H-%M-%S')
COMMIT_MSG ?= "Auto commit at $(COMMIT_TIME)"

.PHONY: sync
# 默认目标：提交并推送当前分支
sync:
	@echo "🚀 开始处理分支: $(CURRENT_BRANCH)"
	@echo "=== 添加所有更改 ==="
	git add .
	@echo ""
	@echo "=== 提交更改 ==="
	@if git diff --cached --quiet; then \
		echo "没有需要提交的更改"; \
	else \
		git commit -m $(COMMIT_MSG) && echo "提交完成"; \
	fi
	@echo ""
	@echo "=== 推送到远程 ($(CURRENT_BRANCH)) ==="
	git push origin $(CURRENT_BRANCH)
	@echo ""
	@echo "✅ 分支 $(CURRENT_BRANCH) 推送完成！"

