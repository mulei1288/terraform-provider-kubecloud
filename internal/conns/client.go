// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"

	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws/credentials"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/aws/session"
	"gitlab.bingosoft.net/bingokube/aws-sdk-go/service/ec2"
)

// BingoCloudClient 统一客户端管理，支持线程安全的客户端缓存
type BingoCloudClient struct {
	Config  *aws.Config      // AWS SDK 配置
	Session *session.Session // AWS Session

	// 服务客户端缓存（线程安全）
	ec2Client     *ec2.EC2
	ec2ClientLock sync.RWMutex
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
		Config:  cfg,
		Session: sess,
	}, nil
}

// EC2Client 获取或创建 EC2 客户端（线程安全，延迟初始化）
func (c *BingoCloudClient) EC2Client() *ec2.EC2 {
	// 快速路径：如果客户端已存在，直接返回
	c.ec2ClientLock.RLock()
	if c.ec2Client != nil {
		defer c.ec2ClientLock.RUnlock()
		return c.ec2Client
	}
	c.ec2ClientLock.RUnlock()

	// 慢速路径：创建新客户端
	c.ec2ClientLock.Lock()
	defer c.ec2ClientLock.Unlock()

	// 双重检查：防止并发创建
	if c.ec2Client == nil {
		c.ec2Client = ec2.New(c.Session)
	}
	return c.ec2Client
}
