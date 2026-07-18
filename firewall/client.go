package firewall

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"

	"github.com/your-username/fwalizer/config"
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

// GetRules 获取实例当前所有防火墙规则（单次最多 100 条）
func (c *Client) GetRules(instanceID string) ([]*lighthouse.FirewallRuleInfo, uint64, error) {
	req := lighthouse.NewDescribeFirewallRulesRequest()
	req.InstanceId = common.StringPtr(instanceID)
	req.Limit = common.Int64Ptr(100) // API 最大分页数，覆盖默认的 20

	resp, err := c.client.DescribeFirewallRules(req)
	if err != nil {
		return nil, 0, fmt.Errorf("查询防火墙规则失败: %w", err)
	}

	version := uint64(0)
	if resp.Response.FirewallVersion != nil {
		version = *resp.Response.FirewallVersion
	}

	return resp.Response.FirewallRuleSet, version, nil
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
