package portconv

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"单端口", "80", []string{"80"}},
		{"多端口", "80,443", []string{"80", "443"}},
		{"范围", "8000-8010", []string{"8000-8010"}},
		{"混合", "80,443,8000-8010", []string{"80", "443", "8000-8010"}},
		{"ALL", "ALL", []string{"ALL"}},
		{"all小写", "all", []string{"ALL"}},
		{"空字符串", "", []string{"ALL"}},
		{"带空格", " 80 , 443 ", []string{"80", "443"}},
		{"尾部逗号", "80,443,", []string{"80", "443"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestToSlash(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"单端口", "80", "80/80"},
		{"范围", "8000-8010", "8000/8010"},
		{"ALL", "ALL", "-1/-1"},
		{"all小写", "all", "-1/-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToSlash(tt.input)
			if got != tt.want {
				t.Errorf("ToSlash(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
