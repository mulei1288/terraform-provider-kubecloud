// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws/credentials"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws/session"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/service/ec2"
)

// BingoCloudClient 封装 BingoCloud SDK 客户端
type BingoCloudClient struct {
	EC2Client *ec2.EC2    // EC2 服务客户端（虚拟机管理）
	Config    *aws.Config // AWS SDK 配置
}

// NewBingoCloudClient 创建新的 BingoCloud 客户端
func NewBingoCloudClient(endpoint, accessKey, secretKey, region string, insecureSkipTLS bool) (*BingoCloudClient, error) {
	// 配置 AWS SDK
	cfg := &aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}

	// 如果需要跳过 TLS 验证（私有云环境常用）
	if insecureSkipTLS {
		cfg.HTTPClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}

	// 创建 session
	sess, err := session.NewSession(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建 AWS session 失败: %w", err)
	}

	return &BingoCloudClient{
		EC2Client: ec2.New(sess),
		Config:    cfg,
	}, nil
}
