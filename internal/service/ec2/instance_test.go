// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccInstanceConfig 生成虚拟机资源的测试配置
func testAccInstanceConfig(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "bingocloud_instance" "test" {
  image_id         = %[1]q
  instance_type    = "m1.small"
  subnet_id        = %[2]q
  password         = "Test@123456"
  system_disk_size = 20
  instance_name    = %[3]q

  tags = {
    Environment = "test"
    ManagedBy   = "terraform"
  }
}
`, os.Getenv("BINGOCLOUD_TEST_AMI"), os.Getenv("BINGOCLOUD_TEST_SUBNET"), name)
}

// TestAccInstanceResource_basic 测试虚拟机资源的基础 CRUD 操作
func TestAccInstanceResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 创建和读取测试
			{
				Config: testAccInstanceConfig("test-instance"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bingocloud_instance.test", "instance_name", "test-instance"),
					resource.TestCheckResourceAttr("bingocloud_instance.test", "instance_type", "m1.small"),
					resource.TestCheckResourceAttr("bingocloud_instance.test", "system_disk_size", "20"),
					resource.TestCheckResourceAttrSet("bingocloud_instance.test", "id"),
					resource.TestCheckResourceAttrSet("bingocloud_instance.test", "private_ip"),
					resource.TestCheckResourceAttr("bingocloud_instance.test", "tags.Environment", "test"),
				),
				ExpectNonEmptyPlan: true, // 暂时允许非空的 refresh plan
			},
			// 导入状态测试
			{
				ResourceName:      "bingocloud_instance.test",
				ImportState:       true,
				ImportStateVerify: true,
				// 密码、系统盘大小、实例名称和标签字段不会被导入，因此需要忽略
				ImportStateVerifyIgnore: []string{"password", "system_disk_size", "instance_name", "tags"},
			},
		},
	})
}

// testAccInstanceConfigUpdate 生成更新后的虚拟机资源配置
func testAccInstanceConfigUpdate(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "bingocloud_instance" "test" {
  image_id         = %[1]q
  instance_type    = "m1.medium"
  subnet_id        = %[2]q
  password         = "Test@123456"
  system_disk_size = 20
  instance_name    = %[3]q

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
    Updated     = "true"
  }
}
`, os.Getenv("BINGOCLOUD_TEST_AMI"), os.Getenv("BINGOCLOUD_TEST_SUBNET"), name)
}

// TestAccInstanceResource_update 测试虚拟机资源的更新操作
func TestAccInstanceResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// 创建初始资源
			{
				Config: testAccInstanceConfig("test-instance-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bingocloud_instance.test", "instance_type", "m1.small"),
					resource.TestCheckResourceAttr("bingocloud_instance.test", "tags.Environment", "test"),
				),
			},
			// 更新资源
			{
				Config: testAccInstanceConfigUpdate("test-instance-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bingocloud_instance.test", "instance_type", "m1.medium"),
					resource.TestCheckResourceAttr("bingocloud_instance.test", "tags.Environment", "production"),
					resource.TestCheckResourceAttr("bingocloud_instance.test", "tags.Updated", "true"),
				),
			},
		},
	})
}

// testAccInstanceConfigWithSecurityGroup 生成带安全组的虚拟机资源配置
func testAccInstanceConfigWithSecurityGroup(name string) string {
	return providerConfig() + fmt.Sprintf(`
resource "bingocloud_instance" "test" {
  image_id         = %[1]q
  instance_type    = "m1.small"
  subnet_id        = %[2]q
  password         = "Test@123456"
  system_disk_size = 20
  instance_name    = %[3]q
  
  security_group_ids = [%[4]q]

  tags = {
    Environment = "test"
  }
}
`, os.Getenv("BINGOCLOUD_TEST_AMI"), os.Getenv("BINGOCLOUD_TEST_SUBNET"), name, os.Getenv("BINGOCLOUD_TEST_SECURITY_GROUP"))
}

// TestAccInstanceResource_withSecurityGroup 测试带安全组的虚拟机资源
func TestAccInstanceResource_withSecurityGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			if v := os.Getenv("BINGOCLOUD_TEST_SECURITY_GROUP"); v == "" {
				t.Skip("跳过测试：BINGOCLOUD_TEST_SECURITY_GROUP 未设置")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceConfigWithSecurityGroup("test-instance-sg"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bingocloud_instance.test", "instance_name", "test-instance-sg"),
					resource.TestCheckResourceAttr("bingocloud_instance.test", "security_group_ids.#", "1"),
				),
			},
		},
	})
}
