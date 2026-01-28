// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ServicePackage 定义 EC2 服务包
type ServicePackage struct{}

// FrameworkResources 返回该服务的所有资源
func (p *ServicePackage) FrameworkResources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
		// 未来可以添加更多资源
		// NewVolumeResource,
		// NewSnapshotResource,
	}
}

// FrameworkDataSources 返回该服务的所有数据源
func (p *ServicePackage) FrameworkDataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// 未来添加数据源
	}
}
