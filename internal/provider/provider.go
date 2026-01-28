// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/mulei1288/terraform-provider-bingocloud/internal/conns"
	"github.com/mulei1288/terraform-provider-bingocloud/internal/service/ec2"
)

// Ensure BingoCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &BingoCloudProvider{}
var _ provider.ProviderWithFunctions = &BingoCloudProvider{}
var _ provider.ProviderWithEphemeralResources = &BingoCloudProvider{}
var _ provider.ProviderWithActions = &BingoCloudProvider{}

// BingoCloudProvider defines the provider implementation.
type BingoCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// BingoCloudProviderModel describes the provider data model.
type BingoCloudProviderModel struct {
	Endpoint        types.String `tfsdk:"endpoint"`
	AccessKey       types.String `tfsdk:"access_key"`
	SecretKey       types.String `tfsdk:"secret_key"`
	Region          types.String `tfsdk:"region"`
	InsecureSkipTLS types.Bool   `tfsdk:"insecure_skip_tls"`
}

func (p *BingoCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "bingocloud"
	resp.Version = p.version
}

func (p *BingoCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "BingoCloud 私有云 Provider",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "BingoCloud API 端点地址",
				Optional:            true,
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "访问密钥 ID",
				Optional:            true,
				Sensitive:           true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "访问密钥",
				Optional:            true,
				Sensitive:           true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "区域名称",
				Optional:            true,
			},
			"insecure_skip_tls": schema.BoolAttribute{
				MarkdownDescription: "跳过 TLS 证书验证（仅用于开发环境）",
				Optional:            true,
			},
		},
	}
}

func (p *BingoCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data BingoCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// 从环境变量获取配置（如果配置文件中未提供）
	endpoint := data.Endpoint.ValueString()
	if endpoint == "" {
		endpoint = os.Getenv("AWS_ENDPOINT")
	}

	accessKey := data.AccessKey.ValueString()
	if accessKey == "" {
		accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}

	secretKey := data.SecretKey.ValueString()
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	region := data.Region.ValueString()
	if region == "" {
		region = "default"
	}

	insecureSkipTLS := data.InsecureSkipTLS.ValueBool()

	// 验证必需配置
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"缺少 Endpoint 配置",
			"必须通过 provider 配置或 AWS_ENDPOINT 环境变量提供 endpoint",
		)
	}
	if accessKey == "" {
		resp.Diagnostics.AddError(
			"缺少 AccessKey 配置",
			"必须通过 provider 配置或 AWS_ACCESS_KEY_ID 环境变量提供 access_key",
		)
	}
	if secretKey == "" {
		resp.Diagnostics.AddError(
			"缺少 SecretKey 配置",
			"必须通过 provider 配置或 AWS_SECRET_ACCESS_KEY 环境变量提供 secret_key",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// 创建 BingoCloud 客户端
	client, err := conns.NewBingoCloudClient(endpoint, accessKey, secretKey, region, insecureSkipTLS)
	if err != nil {
		resp.Diagnostics.AddError(
			"无法创建 BingoCloud 客户端",
			"创建客户端时发生错误: "+err.Error(),
		)
		return
	}

	// 将客户端传递给资源和数据源
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *BingoCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	// 从 Service Package 自动收集资源
	ec2Service := &ec2.ServicePackage{}
	resources := ec2Service.FrameworkResources(ctx)

	return resources
}

func (p *BingoCloudProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *BingoCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *BingoCloudProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *BingoCloudProvider) Actions(ctx context.Context) []func() action.Action {
	return []func() action.Action{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &BingoCloudProvider{
			version: version,
		}
	}
}
