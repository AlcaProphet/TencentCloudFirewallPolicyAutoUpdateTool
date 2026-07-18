package dns

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Resolver DNS 解析器，使用自定义上游 DNS 服务器
type Resolver struct {
	resolver *net.Resolver
}

// New 创建指定 DNS 服务器的解析器
// server 格式: "8.8.8.8:53"
func New(server string) *Resolver {
	return &Resolver{
		resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 10 * time.Second}
				// 尊重 Go Resolver 传入的 network（UDP 或 TCP），
				// 当 DNS 响应超过 512 字节时 Go 会自动通过 TCP 重试
				return d.DialContext(ctx, network, server)
			},
		},
	}
}

// ResolvedIP 解析结果
type ResolvedIP struct {
	Address string // IP 地址
	IsIPv6  bool   // 是否为 IPv6
}

// Lookup 解析域名的 IPv4 和 IPv6 地址
func (r *Resolver) Lookup(host string) ([]ResolvedIP, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ips, err := r.resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("DNS 解析 %s 失败: %w", host, err)
	}

	var results []ResolvedIP
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			results = append(results, ResolvedIP{
				Address: ipv4.String(),
				IsIPv6:  false,
			})
		} else {
			results = append(results, ResolvedIP{
				Address: ip.String(),
				IsIPv6:  true,
			})
		}
	}

	return results, nil
}
