# BingoCloud Terraform Provider 测试环境配置

## 环境变量设置

在运行验收测试之前，需要设置以下环境变量：

### 必需的环境变量

```bash
# BingoCloud API 端点
export AWS_ENDPOINT="http://10.16.203.4:8663"

# 访问密钥 ID
export AWS_ACCESS_KEY_ID="your-access-key-id"

# 访问密钥
export AWS_SECRET_ACCESS_KEY="your-secret-access-key"
```

### 测试资源环境变量（可选）

```bash
# 测试用镜像 ID
export BINGOCLOUD_TEST_AMI="ami-xxxxxxxx"

# 测试用子网 ID
export BINGOCLOUD_TEST_SUBNET="subnet-xxxxxxxx"

# 测试用安全组 ID（可选）
export BINGOCLOUD_TEST_SECURITY_GROUP="sg-xxxxxxxx"
```

## 运行测试

### 运行单元测试

```bash
go test -v ./internal/provider/
```

### 运行验收测试

```bash
# 设置环境变量后运行验收测试
TF_ACC=1 go test -v ./internal/provider/ -timeout 120m
```

### 运行特定测试

```bash
# 运行基础 CRUD 测试
TF_ACC=1 go test -v ./internal/provider/ -run TestAccInstanceResource_basic

# 运行更新测试
TF_ACC=1 go test -v ./internal/provider/ -run TestAccInstanceResource_update
```

## 注意事项

1. **验收测试会创建真实资源**：验收测试会在 BingoCloud 环境中创建真实的虚拟机实例，测试完成后会自动清理。

2. **测试超时设置**：由于需要等待虚拟机启动，建议设置较长的超时时间（120分钟）。

3. **环境变量检查**：测试会在运行前检查必需的环境变量，如果缺失会报错。

4. **成本考虑**：验收测试会产生实际的资源使用成本，请在测试环境中运行。
