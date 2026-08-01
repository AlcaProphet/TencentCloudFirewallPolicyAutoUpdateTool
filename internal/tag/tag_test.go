package tag

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		comment string
		want    string
	}{
		{"有备注", "auto-dns", "生产API", "[auto-dns] 生产API"},
		{"无备注", "auto-dns", "", "[auto-dns]"},
		{"自定义TAG", "my-tag", "测试", "[my-tag] 测试"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.tag, tt.comment)
			if got != tt.want {
				t.Errorf("Format(%q, %q) = %q, want %q", tt.tag, tt.comment, got, tt.want)
			}
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tests := []struct {
		name        string
		description string
		tag         string
		want        bool
	}{
		{"匹配", "[auto-dns] 生产API", "auto-dns", true},
		{"不匹配", "[other] 规则", "auto-dns", false},
		{"空前缀", "普通规则", "auto-dns", false},
		{"仅TAG", "[auto-dns]", "auto-dns", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPrefix(tt.description, tt.tag)
			if got != tt.want {
				t.Errorf("HasPrefix(%q, %q) = %v, want %v", tt.description, tt.tag, got, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		description string
		tag         string
		wantComment string
		wantOk      bool
	}{
		{"有备注", "[auto-dns] 生产API", "auto-dns", "生产API", true},
		{"无备注", "[auto-dns]", "auto-dns", "", true},
		{"不匹配", "[other] 规则", "auto-dns", "", false},
		{"多空格", "[auto-dns]   备注  ", "auto-dns", "备注", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment, ok := Parse(tt.description, tt.tag)
			if comment != tt.wantComment || ok != tt.wantOk {
				t.Errorf("Parse(%q, %q) = (%q, %v), want (%q, %v)",
					tt.description, tt.tag, comment, ok, tt.wantComment, tt.wantOk)
			}
		})
	}
}
