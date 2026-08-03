package api

import (
	"strings"
	"testing"
)

// 期望的 cloud_type 键（与 config.CloudType 四种类型一致）
var zoneCloudTypes = []string{"tc_lighthouse", "tc_cvm", "ali_swas", "ali_ecs"}

// 腾讯云地域 ID 前缀（腾讯云系列均以 ap-/me-/sa-/na-/eu- 开头）
var tcIDPrefixes = []string{"ap-", "me-", "sa-", "na-", "eu-"}

// 阿里云地域 ID 前缀
var aliIDPrefixes = []string{"cn-", "ap-", "us-", "eu-", "me-", "na-", "sa-"}

func hasPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// TestZoneDataComplete 四个平台均提供地域数据，且地域/可用区 ID 非空
func TestZoneDataComplete(t *testing.T) {
	for _, ct := range zoneCloudTypes {
		regions, ok := zoneData[ct]
		if !ok || len(regions) == 0 {
			t.Fatalf("cloud_type %q 缺少地域数据", ct)
		}
		for _, r := range regions {
			if r.ID == "" {
				t.Errorf("%s: 地域 ID 为空（Name=%q）", ct, r.Name)
			}
			if r.Name == "" {
				t.Errorf("%s: 地域名称为空（ID=%q）", ct, r.ID)
			}
			for _, z := range r.Zones {
				if z == "" {
					t.Errorf("%s/%s: 可用区 ID 为空", ct, r.ID)
				}
			}
		}
	}
}

// TestZoneIDPrefix 地域 ID 前缀符合各云厂商命名规则
func TestZoneIDPrefix(t *testing.T) {
	for _, ct := range zoneCloudTypes {
		prefixes := aliIDPrefixes
		if strings.HasPrefix(ct, "tc_") {
			prefixes = tcIDPrefixes
		}
		for _, r := range zoneData[ct] {
			if !hasPrefix(r.ID, prefixes) {
				t.Errorf("%s: 地域 ID %q 前缀不符合规则", ct, r.ID)
			}
		}
	}
}

// TestZoneNoDuplicate 同一平台内地域 ID 不重复
func TestZoneNoDuplicate(t *testing.T) {
	for _, ct := range zoneCloudTypes {
		seen := make(map[string]bool)
		for _, r := range zoneData[ct] {
			if seen[r.ID] {
				t.Errorf("%s: 地域 ID %q 重复", ct, r.ID)
			}
			seen[r.ID] = true
		}
	}
}
