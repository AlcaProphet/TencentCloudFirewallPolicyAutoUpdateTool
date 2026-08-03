package api

import "net/http"

// ZoneRegion 一个地域及其可用区
type ZoneRegion struct {
	ID    string   `json:"id"`    // 地域 ID（RegionId）
	Name  string   `json:"name"`  // 中文地域名
	Zones []string `json:"zones"` // 可用区 ID 列表（SWAS/Lighthouse 无可用区选择，为空）
}

// zoneData 各平台地域与可用区数据（key = cloud_type）
// 数据来源：PlatformAPIDocs/PlatformZoneGuide/ 下 4 份可用区文档，文档更新时需同步本文件。
// - tc_cvm / ali_ecs：文档含完整地域 ID 与可用区 ID
// - tc_lighthouse：文档仅含中文名，地域/可用区 ID 由 CVM 文档对应推断
// - ali_swas：文档仅含中文地域名，地域 ID 由 ECS 文档对应推断；SWAS 无可用区概念
var zoneData = map[string][]ZoneRegion{
	// ── 腾讯云轻量云（腾讯云Lighthouse可用区.md）──
	"tc_lighthouse": {
		{ID: "ap-beijing", Name: "北京", Zones: []string{"ap-beijing-3", "ap-beijing-6", "ap-beijing-7"}},
		{ID: "ap-guangzhou", Name: "广州", Zones: []string{"ap-guangzhou-3", "ap-guangzhou-4", "ap-guangzhou-6", "ap-guangzhou-7"}},
		{ID: "ap-shanghai", Name: "上海", Zones: []string{"ap-shanghai-2", "ap-shanghai-4", "ap-shanghai-5", "ap-shanghai-8"}},
		{ID: "ap-nanjing", Name: "南京", Zones: []string{"ap-nanjing-1", "ap-nanjing-2", "ap-nanjing-3"}},
		{ID: "ap-chengdu", Name: "成都", Zones: []string{"ap-chengdu-1", "ap-chengdu-2"}},
		{ID: "ap-hongkong", Name: "中国香港", Zones: []string{"ap-hongkong-1", "ap-hongkong-2", "ap-hongkong-3"}},
		{ID: "ap-singapore", Name: "新加坡", Zones: []string{"ap-singapore-1", "ap-singapore-2", "ap-singapore-3", "ap-singapore-4"}},
		{ID: "ap-tokyo", Name: "东京", Zones: []string{"ap-tokyo-1", "ap-tokyo-2"}},
		{ID: "ap-jakarta", Name: "雅加达", Zones: []string{"ap-jakarta-1", "ap-jakarta-2"}},
		{ID: "ap-seoul", Name: "首尔", Zones: []string{"ap-seoul-1", "ap-seoul-2"}},
		{ID: "na-siliconvalley", Name: "硅谷", Zones: []string{"na-siliconvalley-1", "na-siliconvalley-2"}},
		{ID: "na-ashburn", Name: "弗吉尼亚", Zones: []string{"na-ashburn-1", "na-ashburn-2"}},
		{ID: "eu-frankfurt", Name: "法兰克福", Zones: []string{"eu-frankfurt-1", "eu-frankfurt-2"}},
		{ID: "sa-saopaulo", Name: "圣保罗", Zones: []string{"sa-saopaulo-1"}},
		{ID: "ap-bangkok", Name: "曼谷", Zones: []string{"ap-bangkok-1", "ap-bangkok-2"}},
	},
	// ── 腾讯云 CVM（腾讯云CVM可用区.md）──
	"tc_cvm": {
		{ID: "ap-beijing", Name: "华北地区（北京）", Zones: []string{"ap-beijing-3", "ap-beijing-4", "ap-beijing-5", "ap-beijing-6", "ap-beijing-7", "ap-beijing-8"}},
		{ID: "ap-shanghai", Name: "华东地区（上海）", Zones: []string{"ap-shanghai-2", "ap-shanghai-3", "ap-shanghai-4", "ap-shanghai-5", "ap-shanghai-8", "ap-shanghai-9"}},
		{ID: "ap-nanjing", Name: "华东地区（南京）", Zones: []string{"ap-nanjing-1", "ap-nanjing-2", "ap-nanjing-3"}},
		{ID: "ap-guangzhou", Name: "华南地区（广州）", Zones: []string{"ap-guangzhou-3", "ap-guangzhou-4", "ap-guangzhou-5", "ap-guangzhou-6", "ap-guangzhou-7"}},
		{ID: "ap-chengdu", Name: "西南地区（成都）", Zones: []string{"ap-chengdu-1", "ap-chengdu-2"}},
		{ID: "ap-chongqing", Name: "西南地区（重庆）", Zones: []string{"ap-chongqing-1"}},
		{ID: "ap-zhongwei", Name: "西北地区（中卫）", Zones: []string{"ap-zhongwei-1"}},
		{ID: "ap-hongkong", Name: "港澳台地区（中国香港）", Zones: []string{"ap-hongkong-1", "ap-hongkong-2", "ap-hongkong-3"}},
		{ID: "ap-beijing-fsi", Name: "北京金融（仅限金融机构和企业申请开通）", Zones: []string{"ap-beijing-fsi-1", "ap-beijing-fsi-2"}},
		{ID: "ap-shanghai-fsi", Name: "上海金融（仅限金融机构和企业申请开通）", Zones: []string{"ap-shanghai-fsi-1", "ap-shanghai-fsi-2", "ap-shanghai-fsi-3", "ap-shanghai-fsi-4"}},
		{ID: "ap-shenzhen-fsi", Name: "深圳金融（仅限金融机构和企业申请开通）", Zones: []string{"ap-shenzhen-fsi-1", "ap-shenzhen-fsi-2", "ap-shenzhen-fsi-3"}},
		{ID: "ap-shanghai-adc", Name: "上海自动驾驶云", Zones: []string{"ap-shanghai-adc-1", "ap-shanghai-adc-2", "ap-shanghai-adc-3", "ap-shanghai-adc-4"}},
		{ID: "ap-singapore", Name: "亚太和中东（新加坡）", Zones: []string{"ap-singapore-1", "ap-singapore-2", "ap-singapore-3", "ap-singapore-4"}},
		{ID: "ap-jakarta", Name: "亚太和中东（雅加达）", Zones: []string{"ap-jakarta-1", "ap-jakarta-2", "ap-jakarta-3"}},
		{ID: "ap-seoul", Name: "亚太和中东（首尔）", Zones: []string{"ap-seoul-1", "ap-seoul-2"}},
		{ID: "ap-tokyo", Name: "亚太和中东（东京）", Zones: []string{"ap-tokyo-1", "ap-tokyo-2"}},
		{ID: "ap-bangkok", Name: "亚太和中东（曼谷）", Zones: []string{"ap-bangkok-1", "ap-bangkok-2"}},
		{ID: "me-saudi-arabia", Name: "亚太和中东（沙特阿拉伯）", Zones: []string{"me-saudi-arabia-1", "me-saudi-arabia-2"}},
		{ID: "sa-saopaulo", Name: "欧洲和美洲（圣保罗）", Zones: []string{"sa-saopaulo-1"}},
		{ID: "na-siliconvalley", Name: "欧洲和美洲（硅谷）", Zones: []string{"na-siliconvalley-1", "na-siliconvalley-2"}},
		{ID: "na-ashburn", Name: "欧洲和美洲（弗吉尼亚）", Zones: []string{"na-ashburn-1", "na-ashburn-2"}},
		{ID: "eu-frankfurt", Name: "欧洲和美洲（法兰克福）", Zones: []string{"eu-frankfurt-1", "eu-frankfurt-2"}},
	},
	// ── 阿里云轻量云（阿里云SWAS可用区.md，ID 由 ECS 文档推断，SWAS 无可用区选择）──
	"ali_swas": {
		{ID: "cn-wulanchabu", Name: "华北6（乌兰察布）"},
		{ID: "cn-heyuan", Name: "华南2（河源）"},
		{ID: "cn-beijing", Name: "华北2（北京）"},
		{ID: "cn-shanghai", Name: "华东2（上海）"},
		{ID: "cn-hangzhou", Name: "华东1（杭州）"},
		{ID: "cn-shenzhen", Name: "华南1（深圳）"},
		{ID: "cn-guangzhou", Name: "华南3（广州）"},
		{ID: "cn-chengdu", Name: "西南1（成都）"},
		{ID: "cn-wuhan-lr", Name: "华中1（武汉-本地地域）"},
		{ID: "cn-nanjing", Name: "华东5（南京-本地地域）"},
		{ID: "cn-qingdao", Name: "华北1（青岛）"},
		{ID: "cn-fuzhou", Name: "华东6（福州-本地地域）"},
		{ID: "cn-zhangjiakou", Name: "华北3（张家口）"},
		{ID: "cn-huhehaote", Name: "华北5（呼和浩特）"},
		{ID: "cn-hongkong", Name: "中国香港"},
		{ID: "ap-southeast-1", Name: "新加坡"},
		{ID: "ap-southeast-3", Name: "马来西亚（吉隆坡）"},
		{ID: "ap-southeast-5", Name: "印度尼西亚（雅加达）"},
		{ID: "ap-southeast-6", Name: "菲律宾（马尼拉）"},
		{ID: "ap-southeast-7", Name: "泰国（曼谷）"},
		{ID: "ap-northeast-1", Name: "日本（东京）"},
		{ID: "ap-northeast-2", Name: "韩国（首尔）"},
		{ID: "eu-west-1", Name: "英国（伦敦）"},
		{ID: "eu-central-1", Name: "德国（法兰克福）"},
		{ID: "us-east-1", Name: "美国（弗吉尼亚）"},
		{ID: "us-west-1", Name: "美国（硅谷）"},
	},
	// ── 阿里云 ECS（阿里云ECS可用区.md）──
	"ali_ecs": {
		{ID: "cn-qingdao", Name: "华北1（青岛）", Zones: []string{"cn-qingdao-b", "cn-qingdao-c"}},
		{ID: "cn-beijing", Name: "华北2（北京）", Zones: []string{"cn-beijing-a", "cn-beijing-b", "cn-beijing-c", "cn-beijing-d", "cn-beijing-e", "cn-beijing-f", "cn-beijing-g", "cn-beijing-h", "cn-beijing-i", "cn-beijing-j", "cn-beijing-k", "cn-beijing-l"}},
		{ID: "cn-zhangjiakou", Name: "华北3（张家口）", Zones: []string{"cn-zhangjiakou-a", "cn-zhangjiakou-b", "cn-zhangjiakou-c"}},
		{ID: "cn-huhehaote", Name: "华北5（呼和浩特）", Zones: []string{"cn-huhehaote-a", "cn-huhehaote-b"}},
		{ID: "cn-wulanchabu", Name: "华北6（乌兰察布）", Zones: []string{"cn-wulanchabu-a", "cn-wulanchabu-b", "cn-wulanchabu-c", "cn-wulanchabu-d"}},
		{ID: "cn-hangzhou", Name: "华东1（杭州）", Zones: []string{"cn-hangzhou-b", "cn-hangzhou-e", "cn-hangzhou-f", "cn-hangzhou-g", "cn-hangzhou-h", "cn-hangzhou-i", "cn-hangzhou-j", "cn-hangzhou-k"}},
		{ID: "cn-shanghai", Name: "华东2（上海）", Zones: []string{"cn-shanghai-a", "cn-shanghai-b", "cn-shanghai-c", "cn-shanghai-d", "cn-shanghai-e", "cn-shanghai-f", "cn-shanghai-g", "cn-shanghai-k", "cn-shanghai-l", "cn-shanghai-m", "cn-shanghai-n", "cn-shanghai-o"}},
		{ID: "cn-nanjing", Name: "华东5（南京-本地地域-关停中）", Zones: []string{"cn-nanjing-a"}},
		{ID: "cn-fuzhou", Name: "华东6（福州-本地地域-关停中）", Zones: []string{"cn-fuzhou-a"}},
		{ID: "cn-wuhan-lr", Name: "华中1（武汉-本地地域）", Zones: []string{"cn-wuhan-lr-a"}},
		{ID: "cn-shenzhen", Name: "华南1（深圳）", Zones: []string{"cn-shenzhen-a", "cn-shenzhen-b", "cn-shenzhen-c", "cn-shenzhen-d", "cn-shenzhen-e", "cn-shenzhen-f"}},
		{ID: "cn-heyuan", Name: "华南2（河源）", Zones: []string{"cn-heyuan-a", "cn-heyuan-b", "cn-heyuan-c"}},
		{ID: "cn-guangzhou", Name: "华南3（广州）", Zones: []string{"cn-guangzhou-a", "cn-guangzhou-b"}},
		{ID: "cn-chengdu", Name: "西南1（成都）", Zones: []string{"cn-chengdu-a", "cn-chengdu-b", "cn-chengdu-c"}},
		{ID: "cn-zhongwei", Name: "西北2（中卫）", Zones: []string{"cn-zhongwei-a", "cn-zhongwei-b"}},
		{ID: "cn-hongkong", Name: "中国香港", Zones: []string{"cn-hongkong-b", "cn-hongkong-c", "cn-hongkong-d"}},
		{ID: "ap-southeast-1", Name: "新加坡", Zones: []string{"ap-southeast-1a", "ap-southeast-1b", "ap-southeast-1c", "ap-southeast-1d"}},
		{ID: "ap-southeast-3", Name: "马来西亚（吉隆坡）", Zones: []string{"ap-southeast-3a", "ap-southeast-3b", "ap-southeast-3c"}},
		{ID: "ap-southeast-5", Name: "印度尼西亚（雅加达）", Zones: []string{"ap-southeast-5a", "ap-southeast-5b", "ap-southeast-5c"}},
		{ID: "ap-southeast-6", Name: "菲律宾（马尼拉）", Zones: []string{"ap-southeast-6a", "ap-southeast-6b"}},
		{ID: "ap-southeast-7", Name: "泰国（曼谷）", Zones: []string{"ap-southeast-7a", "ap-southeast-7b"}},
		{ID: "ap-southeast-8", Name: "马来西亚（柔佛州）", Zones: []string{"ap-southeast-8a", "ap-southeast-8b"}},
		{ID: "ap-northeast-1", Name: "日本（东京）", Zones: []string{"ap-northeast-1a", "ap-northeast-1b", "ap-northeast-1c", "ap-northeast-1d", "ap-northeast-1e"}},
		{ID: "ap-northeast-2", Name: "韩国（首尔）", Zones: []string{"ap-northeast-2a", "ap-northeast-2b", "ap-northeast-2c"}},
		{ID: "us-west-1", Name: "美国（硅谷）", Zones: []string{"us-west-1a", "us-west-1b"}},
		{ID: "us-east-1", Name: "美国（弗吉尼亚）", Zones: []string{"us-east-1a", "us-east-1b"}},
		{ID: "eu-central-1", Name: "德国（法兰克福）", Zones: []string{"eu-central-1a", "eu-central-1b", "eu-central-1c"}},
		{ID: "eu-west-1", Name: "英国（伦敦）", Zones: []string{"eu-west-1a", "eu-west-1b"}},
		{ID: "eu-west-2", Name: "法国（巴黎）", Zones: []string{"eu-west-2a", "eu-west-2b"}},
		{ID: "me-east-1", Name: "阿联酋（迪拜）", Zones: []string{"me-east-1a", "me-east-1b"}},
		{ID: "me-central-1", Name: "沙特（利雅得）- 合作伙伴运营", Zones: []string{"me-central-1a", "me-central-1b"}},
		{ID: "na-south-1", Name: "墨西哥", Zones: []string{"na-south-1a", "na-south-1b"}},
		{ID: "sa-east-1", Name: "巴西（圣保罗）", Zones: []string{"sa-east-1a", "sa-east-1b"}},
	},
}

// handleGetZones 返回各平台地域与可用区数据（前端地域自动补全数据源）
func (d *Deps) handleGetZones(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, zoneData)
}
