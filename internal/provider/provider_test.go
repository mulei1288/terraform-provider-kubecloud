// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories 用于验收测试的 Provider 工厂
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bingocloud": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck 验收测试前置检查，确保必需的环境变量已设置
func testAccPreCheck(t *testing.T) {
	// 检查必需的环境变量
	if v := os.Getenv("AWS_ENDPOINT"); v == "" {
		t.Fatal("AWS_ENDPOINT 环境变量必须设置用于验收测试")
	}
	if v := os.Getenv("AWS_ACCESS_KEY_ID"); v == "" {
		t.Fatal("AWS_ACCESS_KEY_ID 环境变量必须设置用于验收测试")
	}
	if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v == "" {
		t.Fatal("AWS_SECRET_ACCESS_KEY 环境变量必须设置用于验收测试")
	}
}

// providerConfig 返回基础的 Provider 配置
func providerConfig() string {
	return `
provider "bingocloud" {
  # 配置通过环境变量提供：
  # AWS_ENDPOINT
  # AWS_ACCESS_KEY_ID
  # AWS_SECRET_ACCESS_KEY
}
`
}
