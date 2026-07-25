package portconv

import "strings"

// Parse 将统一端口格式拆分为独立端口/范围列表
// 输入: "80,443,8000-8010" → ["80", "443", "8000-8010"]
// 输入: "ALL" → ["ALL"]
func Parse(ports string) []string {
	ports = strings.TrimSpace(ports)
	if ports == "" || strings.EqualFold(ports, "ALL") {
		return []string{"ALL"}
	}
	parts := strings.Split(ports, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// ToSlash 将单个端口/范围转为阿里云斜杠格式
// "80" → "80/80"，"8000-8010" → "8000/8010"，"ALL" → "-1/-1"
func ToSlash(port string) string {
	if strings.EqualFold(port, "ALL") {
		return "-1/-1"
	}
	if strings.Contains(port, "-") {
		parts := strings.SplitN(port, "-", 2)
		return parts[0] + "/" + parts[1]
	}
	return port + "/" + port
}
