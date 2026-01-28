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
