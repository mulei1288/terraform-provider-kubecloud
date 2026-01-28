# BingoCloud Instance 资源示例

这个示例展示了如何使用 BingoCloud Terraform Provider 创建和管理虚拟机实例。

## 前置条件

1. 已安装 Terraform（版本 >= 1.0）
2. 有可用的 BingoCloud 环境和访问凭证
3. 已编译 terraform-provider-bingocloud

## 使用步骤

### 1. 编译 Provider

在项目根目录下编译 provider：

```bash
cd /Users/pengzz/go/src/github.com/mulei1288/terraform-provider-bingocloud
go build -o terraform-provider-bingocloud
```

### 2. 配置本地 Provider

创建 Terraform 本地 provider 配置：

```bash
# 创建本地 provider 目录
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/mulei1288/bingocloud/0.1.0/darwin_arm64

# 复制编译好的 provider
cp terraform-provider-bingocloud ~/.terraform.d/plugins/registry.terraform.io/mulei1288/bingocloud/0.1.0/darwin_arm64/
```


### 3. 修改配置文件

编辑 `resource.tf` 文件，根据你的实际环境修改以下参数：

- `endpoint`: BingoCloud API 端点地址
- `access_key`: 你的访问密钥 ID
- `secret_key`: 你的访问密钥
- `ami`: 可用的镜像 ID
- `subnet_id`: 可用的子网 ID
- `password`: 虚拟机登录密码（需符合复杂度要求）

### 4. 初始化 Terraform

```bash
cd examples/resources/bingocloud_instance
terraform init
```

### 5. 查看执行计划

```bash
terraform plan
```

### 6. 创建资源

```bash
terraform apply
```

输入 `yes` 确认创建。

### 7. 查看资源状态

```bash
terraform show
```

### 8. 销毁资源

测试完成后，清理资源：

```bash
terraform destroy
```

输入 `yes` 确认销毁。


## 资源配置说明

### 必需参数

- `ami` - 镜像 ID
- `instance_type` - 实例类型（如 m1.small, m1.medium）
- `subnet_id` - 子网 ID
- `password` - 登录密码

### 可选参数

- `name` - 实例名称
- `system_disk_size` - 系统盘大小（GB）
- `security_group_ids` - 安全组 ID 列表
- `tags` - 标签键值对

## 输出值

- `instance_id` - 创建的实例 ID
- `instance_private_ip` - 实例的私有 IP 地址

## 注意事项

1. 密码必须符合复杂度要求（包含大小写字母、数字和特殊字符）
2. 创建实例需要一定时间，请耐心等待
3. 测试完成后记得销毁资源，避免产生不必要的费用

## 故障排查

如果遇到问题，可以设置环境变量查看详细日志：

```bash
export TF_LOG=DEBUG
terraform apply
```
