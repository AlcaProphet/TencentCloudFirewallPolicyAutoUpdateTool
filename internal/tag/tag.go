package tag

import (
	"fmt"
	"strings"
)

// Format 生成规则描述："[TAG] comment"
func Format(tag, comment string) string {
	if comment == "" {
		return fmt.Sprintf("[%s]", tag)
	}
	return fmt.Sprintf("[%s] %s", tag, comment)
}

// HasPrefix 判断规则描述是否属于本工具管理
func HasPrefix(description, tag string) bool {
	return strings.HasPrefix(description, "["+tag+"]")
}

// Parse 从描述中提取 comment 部分，非本工具规则返回 ok=false
func Parse(description, tag string) (comment string, ok bool) {
	prefix := "[" + tag + "]"
	if !strings.HasPrefix(description, prefix) {
		return "", false
	}
	return strings.TrimSpace(description[len(prefix):]), true
}
