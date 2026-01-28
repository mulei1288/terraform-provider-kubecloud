# 配置 Terraform 和 Provider
terraform {
  required_providers {
    bingocloud = {
      source = "registry.terraform.io/mulei1288/bingocloud"
    }
  }
}

# 配置 BingoCloud Provider
provider "bingocloud" {
  endpoint            = "http://10.16.203.4"
  access_key          = "19650A325E7A7872E5EA"
  secret_key          = "W0VGMzA4MjAyRkU3QkEyQzA2M0FFQ0ZFRkE2NUI2"
  region              = "cc1"
  insecure_skip_tls   = true
}

# 创建带数据盘的实例
resource "bingocloud_instance" "example" {
  count         = 1
  image_id      = "ami-0FC0EEF6"
  instance_type = "m1.medium"
  subnet_id     = "subnet-A5D2DD8E"
  password      = "bingo@word1"
  instance_name = format("instance-test-%02d", count.index + 1)
  min_count     = 1  # 创建 3 个实例

  block_device_mappings = [
    {
      volume_size = 100
      volume_type = "storage-cloud"
      device_name = "/dev/vda"  # 系统盘
    },
    {
      volume_size = 2
      volume_type = "storage-localfs"
      device_name = "/dev/vdb"  # 数据盘
    },
    {
      volume_size = 100
      volume_type = "storage-cloud"
      device_name = "/dev/vdc"  # 系统盘
    },
  ]

  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

output "instance_info" {
  description = "ECS实例详细信息"
  value = [
    for idx, instance in bingocloud_instance.example : {
      id        = instance.id
      name      = instance.instance_name
      public_ip = instance.public_ip
      private_ip = instance.private_ip
    }
  ]
  sensitive = false
}
