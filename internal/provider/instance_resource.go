// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws/awserr"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/service/ec2"
)

// 确保实现了必需的接口
var _ resource.Resource = &InstanceResource{}
var _ resource.ResourceWithImportState = &InstanceResource{}

// InstanceResource 定义虚拟机资源实现
type InstanceResource struct {
	client *BingoCloudClient
}

// BlockDeviceMappingModel 描述块设备映射配置
type BlockDeviceMappingModel struct {
	VolumeSize types.Int64  `tfsdk:"volume_size"`
	VolumeType types.String `tfsdk:"volume_type"`
	DeviceName types.String `tfsdk:"device_name"`
}

// InstanceResourceModel 描述虚拟机资源数据模型
type InstanceResourceModel struct {
	// 必需参数
	ImageId             types.String `tfsdk:"image_id"`
	InstanceType        types.String `tfsdk:"instance_type"`
	SubnetID            types.String `tfsdk:"subnet_id"`
	Password            types.String `tfsdk:"password"`
	BlockDeviceMappings types.List   `tfsdk:"block_device_mappings"`

	// 可选参数
	MinCount         types.Int64  `tfsdk:"min_count"`
	InstanceName     types.String `tfsdk:"instance_name"`
	SecurityGroupIDs types.List   `tfsdk:"security_group_ids"`
	KeyName          types.String `tfsdk:"key_name"`
	UserData         types.String `tfsdk:"user_data"`
	Tags             types.Map    `tfsdk:"tags"`

	// 计算属性
	ID               types.String `tfsdk:"id"`
	PrivateIP        types.String `tfsdk:"private_ip"`
	PublicIP         types.String `tfsdk:"public_ip"`
	State            types.String `tfsdk:"state"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
}

// NewInstanceResource 创建新的虚拟机资源实例
func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// Metadata 返回资源类型名称
func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Configure 配置资源，接收 Provider 传递的客户端
func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BingoCloudClient)
	if !ok {
		resp.Diagnostics.AddError(
			"意外的资源配置类型",
			fmt.Sprintf("期望 *BingoCloudClient，得到: %T。请向 provider 开发者报告此问题。", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Schema 定义资源的属性架构
func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "管理 BingoCloud 虚拟机实例",

		Attributes: map[string]schema.Attribute{
			// 必需参数
			"image_id": schema.StringAttribute{
				MarkdownDescription: "镜像 ID，用于创建虚拟机",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_type": schema.StringAttribute{
				MarkdownDescription: "实例类型（如 t2.micro, m5.large）",
				Required:            true,
			},
			"subnet_id": schema.StringAttribute{
				MarkdownDescription: "子网 ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "实例登录密码",
				Required:            true,
				Sensitive:           true,
			},
			"block_device_mappings": schema.ListNestedAttribute{
				MarkdownDescription: "块设备映射配置列表，第一个元素为系统盘，后续为数据盘",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"volume_size": schema.Int64Attribute{
							MarkdownDescription: "磁盘大小（GB）",
							Required:            true,
						},
						"volume_type": schema.StringAttribute{
							MarkdownDescription: "磁盘类型（如 gp2, io1）",
							Required:            true,
						},
						"device_name": schema.StringAttribute{
							MarkdownDescription: "设备名称（如 /dev/vda, /dev/vdb），第一个元素默认 /dev/vda",
							Optional:            true,
							Computed:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},

			// 可选参数
			"min_count": schema.Int64Attribute{
				MarkdownDescription: "创建的实例个数，默认为 1",
				Optional:            true,
				Computed:            true,
			},
			"instance_name": schema.StringAttribute{
				MarkdownDescription: "实例名称",
				Optional:            true,
				Computed:            true,
			},
			"security_group_ids": schema.ListAttribute{
				MarkdownDescription: "安全组 ID 列表",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"key_name": schema.StringAttribute{
				MarkdownDescription: "SSH 密钥对名称",
				Optional:            true,
			},
			"user_data": schema.StringAttribute{
				MarkdownDescription: "用户数据脚本（Base64 编码）",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.MapAttribute{
				MarkdownDescription: "资源标签",
				ElementType:         types.StringType,
				Optional:            true,
			},

			// 计算属性（只读）
			"id": schema.StringAttribute{
				MarkdownDescription: "实例 ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_ip": schema.StringAttribute{
				MarkdownDescription: "私有 IP 地址",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "公网 IP 地址",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "实例状态（running, stopped, terminated 等）",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"availability_zone": schema.StringAttribute{
				MarkdownDescription: "可用区",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create 创建虚拟机实例
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan InstanceResourceModel

	// 读取 Terraform 计划数据
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 构建 RunInstances 请求
	instanceCount := int64(1) // 默认创建 1 个实例
	if !plan.MinCount.IsNull() {
		instanceCount = plan.MinCount.ValueInt64()
	} else {
		plan.MinCount = types.Int64Value(instanceCount)
	}

	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(plan.ImageId.ValueString()),
		InstanceType: aws.String(plan.InstanceType.ValueString()),
		SubnetId:     aws.String(plan.SubnetID.ValueString()),
		MinCount:     aws.Int64(instanceCount),
		MaxCount:     aws.Int64(instanceCount),
		InstanceName: aws.String(plan.InstanceName.ValueString()),
		Password:     aws.String(plan.Password.ValueString()),
	}

	// TODO: Password 参数暂时无法设置
	// BingoCloud SDK 的 RunInstancesInput 结构体中没有 Password 字段
	// 需要等待 SDK 更新或找到其他方式传递密码参数

	// 配置块设备映射
	var blockDeviceMappings []*ec2.BlockDeviceMapping
	var bdmList []BlockDeviceMappingModel
	resp.Diagnostics.Append(plan.BlockDeviceMappings.ElementsAs(ctx, &bdmList, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 处理每个块设备映射
	for i, bdm := range bdmList {
		deviceName := bdm.DeviceName.ValueString()

		// 第一个元素（系统盘）如果没有指定 device_name，使用默认值
		if i == 0 && deviceName == "" {
			deviceName = "/dev/vda"
			bdmList[i].DeviceName = types.StringValue(deviceName)
		}

		blockDeviceMappings = append(blockDeviceMappings, &ec2.BlockDeviceMapping{
			DeviceName: aws.String(deviceName),
			Ebs: &ec2.EbsBlockDevice{
				VolumeSize:          aws.Int64(bdm.VolumeSize.ValueInt64()),
				VolumeType:          aws.String(bdm.VolumeType.ValueString()),
				DeleteOnTermination: aws.Bool(true),
			},
		})
	}

	// 更新 plan 中的 BlockDeviceMappings（包含默认值）
	updatedBdmList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"volume_size": types.Int64Type,
			"volume_type": types.StringType,
			"device_name": types.StringType,
		},
	}, bdmList)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		plan.BlockDeviceMappings = updatedBdmList
	}

	runInput.BlockDeviceMappings = blockDeviceMappings

	// 配置安全组
	if !plan.SecurityGroupIDs.IsNull() && !plan.SecurityGroupIDs.IsUnknown() {
		var sgIDs []string
		resp.Diagnostics.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &sgIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		runInput.SecurityGroupIds = aws.StringSlice(sgIDs)
	}

	// 配置密钥对
	if !plan.KeyName.IsNull() {
		runInput.KeyName = aws.String(plan.KeyName.ValueString())
	}

	// 配置用户数据
	if !plan.UserData.IsNull() {
		runInput.UserData = aws.String(plan.UserData.ValueString())
	}

	// 配置标签
	tags := []*ec2.Tag{}
	if !plan.InstanceName.IsNull() {
		tags = append(tags, &ec2.Tag{
			Key:   aws.String("Name"),
			Value: aws.String(plan.InstanceName.ValueString()),
		})
	}
	if !plan.Tags.IsNull() {
		var tagMap map[string]string
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagMap, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range tagMap {
			tags = append(tags, &ec2.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
	}
	if len(tags) > 0 {
		runInput.TagSpecifications = []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags:         tags,
			},
		}
	}

	// 调用 API 创建实例
	tflog.Debug(ctx, "创建 BingoCloud 实例", map[string]interface{}{
		"image_id":      plan.ImageId.ValueString(),
		"instance_type": plan.InstanceType.ValueString(),
	})

	result, err := r.client.EC2Client.RunInstancesWithContext(ctx, runInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"创建实例失败",
			"无法创建实例: "+err.Error(),
		)
		return
	}

	if len(result.Instances) == 0 {
		resp.Diagnostics.AddError(
			"创建实例失败",
			"API 返回空实例列表",
		)
		return
	}

	instance := result.Instances[0]
	plan.ID = types.StringValue(aws.StringValue(instance.InstanceId))

	// 等待实例运行
	tflog.Debug(ctx, "等待实例运行", map[string]interface{}{
		"instance_id": plan.ID.ValueString(),
	})

	err = r.client.EC2Client.WaitUntilInstanceRunningWithContext(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"等待实例运行失败",
			"实例创建成功但未能进入运行状态: "+err.Error(),
		)
		return
	}

	// 读取实例详细信息以填充计算属性
	describeResult, err := r.client.EC2Client.DescribeInstancesWithContext(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []*string{instance.InstanceId},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"读取实例详情失败",
			"实例创建成功但无法读取详细信息: "+err.Error(),
		)
		return
	}

	if len(describeResult.Reservations) > 0 && len(describeResult.Reservations[0].Instances) > 0 {
		inst := describeResult.Reservations[0].Instances[0]

		// 填充计算属性
		plan.State = types.StringValue(aws.StringValue(inst.State.Name))
		plan.AvailabilityZone = types.StringValue(aws.StringValue(inst.Placement.AvailabilityZone))

		if inst.PrivateIpAddress != nil {
			plan.PrivateIP = types.StringValue(aws.StringValue(inst.PrivateIpAddress))
		}
		if inst.PublicIpAddress != nil {
			plan.PublicIP = types.StringValue(aws.StringValue(inst.PublicIpAddress))
		}

		// 如果用户没有提供 security_group_ids，从 API 读取并填充
		if plan.SecurityGroupIDs.IsNull() || plan.SecurityGroupIDs.IsUnknown() {
			sgIDs := make([]string, 0, len(inst.SecurityGroups))
			for _, sg := range inst.SecurityGroups {
				if sg.GroupId != nil {
					sgIDs = append(sgIDs, aws.StringValue(sg.GroupId))
				}
			}
			sgList, diags := types.ListValueFrom(ctx, types.StringType, sgIDs)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				plan.SecurityGroupIDs = sgList
			}
		}
	}

	// 读取最新状态并保存
	tflog.Trace(ctx, "创建实例成功")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read 读取虚拟机实例状态
func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state InstanceResourceModel

	// 读取 Terraform 状态数据
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 调用 DescribeInstances API
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(state.ID.ValueString())},
	}

	result, err := r.client.EC2Client.DescribeInstancesWithContext(ctx, input)
	if err != nil {
		if isNotFoundError(err) {
			// 实例不存在，从状态中移除
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"读取实例失败",
			"无法读取实例 "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		// 实例不存在，从状态中移除
		resp.State.RemoveResource(ctx)
		return
	}

	instance := result.Reservations[0].Instances[0]

	// 更新模型 - 基本属性
	state.ImageId = types.StringValue(aws.StringValue(instance.ImageId))
	state.InstanceType = types.StringValue(aws.StringValue(instance.InstanceType))
	state.SubnetID = types.StringValue(aws.StringValue(instance.SubnetId))
	state.State = types.StringValue(aws.StringValue(instance.State.Name))
	state.AvailabilityZone = types.StringValue(aws.StringValue(instance.Placement.AvailabilityZone))

	// IP 地址
	if instance.PrivateIpAddress != nil {
		state.PrivateIP = types.StringValue(aws.StringValue(instance.PrivateIpAddress))
	}
	if instance.PublicIpAddress != nil {
		state.PublicIP = types.StringValue(aws.StringValue(instance.PublicIpAddress))
	}

	// 密钥对 - 只在实例有密钥对时才设置
	if instance.KeyName != nil && aws.StringValue(instance.KeyName) != "" {
		state.KeyName = types.StringValue(aws.StringValue(instance.KeyName))
	}

	// 安全组
	sgIDs := make([]string, 0, len(instance.SecurityGroups))
	for _, sg := range instance.SecurityGroups {
		if sg.GroupId != nil {
			sgIDs = append(sgIDs, aws.StringValue(sg.GroupId))
		}
	}
	sgList, diags := types.ListValueFrom(ctx, types.StringType, sgIDs)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		state.SecurityGroupIDs = sgList
	}

	// 标签
	tagMap := make(map[string]string)
	for _, tag := range instance.Tags {
		if tag.Key != nil && tag.Value != nil {
			key := aws.StringValue(tag.Key)
			if key == "Name" {
				state.InstanceName = types.StringValue(aws.StringValue(tag.Value))
			} else {
				tagMap[key] = aws.StringValue(tag.Value)
			}
		}
	}
	// 始终设置 tags，即使为空
	tagsValue, diags := types.MapValueFrom(ctx, types.StringType, tagMap)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		state.Tags = tagsValue
	}

	// 注意：system_disk_size 在导入时无法从 API 读取，保持状态中的值不变

	// 保存更新后的状态
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update 更新虚拟机实例（暂不支持）
func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan InstanceResourceModel

	// 读取计划数据
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 暂时直接保存状态，后续可以实现具体的更新逻辑
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete 删除虚拟机实例
func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state InstanceResourceModel

	// 读取 Terraform 状态数据
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 调用 TerminateInstances API
	_, err := r.client.EC2Client.TerminateInstancesWithContext(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(state.ID.ValueString())},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"删除实例失败",
			"无法删除实例 "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "删除实例成功")
}

// ImportState 支持通过实例 ID 导入资源
func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// isNotFoundError 判断是否为资源不存在错误
func isNotFoundError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == "InvalidInstanceID.NotFound"
	}
	return false
}
