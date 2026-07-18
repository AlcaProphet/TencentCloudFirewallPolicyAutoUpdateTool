package firewall

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"

	"fwalizer/config"
)

// Client 腾讯云 Lighthouse 防火墙客户端封装
type Client struct {
	client *lighthouse.Client
}

// NewClient 创建 Lighthouse 客户端
func NewClient(cfg *config.Config) (*Client, error) {
	credential := common.NewCredential(cfg.SecretID, cfg.SecretKey)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	cpf.HttpProfile.ReqMethod = "POST"

	client, err := lighthouse.NewClient(credential, cfg.Region, cpf)
	if err != nil {
		return nil, fmt.Errorf("创建 Lighthouse 客户端失败: %w", err)
	}

	return &Client{client: client}, nil
}

// GetRules 获取实例当前所有防火墙规则（自动分页，确保不漏数据）
func (c *Client) GetRules(instanceID string) ([]*lighthouse.FirewallRuleInfo, uint64, error) {
	var allRules []*lighthouse.FirewallRuleInfo
	var version uint64
	offset := int64(0)
	pageSize := int64(100) // API 单次最大返回数

	for {
		req := lighthouse.NewDescribeFirewallRulesRequest()
		req.InstanceId = common.StringPtr(instanceID)
		req.Limit = common.Int64Ptr(pageSize)
		req.Offset = common.Int64Ptr(offset)

		resp, err := c.client.DescribeFirewallRules(req)
		if err != nil {
			return nil, 0, fmt.Errorf("查询防火墙规则失败: %w", err)
		}

		// 记录 FirewallVersion（每页一致，取最后一次即可）
		if resp.Response.FirewallVersion != nil {
			version = *resp.Response.FirewallVersion
		}

		allRules = append(allRules, resp.Response.FirewallRuleSet...)

		// 判断是否已拉完所有规则
		totalCount := int64(0)
		if resp.Response.TotalCount != nil {
			totalCount = *resp.Response.TotalCount
		}

		offset += int64(len(resp.Response.FirewallRuleSet))
		if offset >= totalCount || len(resp.Response.FirewallRuleSet) == 0 {
			break
		}
	}

	return allRules, version, nil
}

// CreateRules 批量添加防火墙规则
func (c *Client) CreateRules(instanceID string, rules []*lighthouse.FirewallRule) error {
	if len(rules) == 0 {
		return nil
	}

	req := lighthouse.NewCreateFirewallRulesRequest()
	req.InstanceId = common.StringPtr(instanceID)
	req.FirewallRules = rules

	_, err := c.client.CreateFirewallRules(req)
	if err != nil {
		return fmt.Errorf("添加防火墙规则失败: %w", err)
	}

	return nil
}

// DeleteRules 批量删除防火墙规则
func (c *Client) DeleteRules(instanceID string, rules []*lighthouse.FirewallRule) error {
	if len(rules) == 0 {
		return nil
	}

	req := lighthouse.NewDeleteFirewallRulesRequest()
	req.InstanceId = common.StringPtr(instanceID)
	req.FirewallRules = rules

	_, err := c.client.DeleteFirewallRules(req)
	if err != nil {
		return fmt.Errorf("删除防火墙规则失败: %w", err)
	}

	return nil
}
