package config

import "fmt"

// Validate 校验配置完整性
func (c *Config) Validate() error {
	if len(c.Targets) == 0 {
		return fmt.Errorf("TARGETS 不能为空")
	}
	if len(c.DomainRules) == 0 {
		return fmt.Errorf("RULES 不能为空")
	}
	// 检查每个 Target 是否有对应凭据
	for i, t := range c.Targets {
		switch t.CloudType {
		case CloudTCLighthouse, CloudTCCVM:
			if c.TCAccessID == "" || c.TCAccessKey == "" {
				return fmt.Errorf("TARGETS[%d] 为腾讯云，但 TC_ACCESS_ID/TC_ACCESS_KEY 未设置", i+1)
			}
		case CloudAliSWAS, CloudAliECS:
			if c.AliAccessID == "" || c.AliAccessKey == "" {
				return fmt.Errorf("TARGETS[%d] 为阿里云，但 ALI_ACCESS_ID/ALI_ACCESS_KEY 未设置", i+1)
			}
		}
		if t.ResourceID == "" {
			return fmt.Errorf("TARGETS[%d] resource_id 不能为空", i+1)
		}
		if t.Region == "" {
			return fmt.Errorf("TARGETS[%d] region 不能为空", i+1)
		}
	}
	return nil
}
