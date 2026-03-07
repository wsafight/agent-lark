package base

import "testing"

func TestParseBitableURL(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantAppToken  string
		wantTableID   string
	}{
		{
			name:         "标准飞书多维表格 URL",
			input:        "https://company.feishu.cn/base/bascnABCDEF?table=tblXYZ123",
			wantAppToken: "bascnABCDEF",
			wantTableID:  "tblXYZ123",
		},
		{
			name:         "URL 带有额外查询参数",
			input:        "https://company.feishu.cn/base/bascnABCDEF?table=tblXYZ123&view=vewAAA",
			wantAppToken: "bascnABCDEF",
			wantTableID:  "tblXYZ123",
		},
		{
			name:         "无 table 参数",
			input:        "https://company.feishu.cn/base/bascnABCDEF",
			wantAppToken: "bascnABCDEF",
			wantTableID:  "",
		},
		{
			name:         "裸 appToken（非 http 开头）",
			input:        "bascnABCDEF",
			wantAppToken: "bascnABCDEF",
			wantTableID:  "",
		},
		{
			name:         "路径中无 base 段",
			input:        "https://company.feishu.cn/wiki/somepage",
			wantAppToken: "",
			wantTableID:  "",
		},
		{
			name:         "空字符串",
			input:        "",
			wantAppToken: "",
			wantTableID:  "",
		},
		{
			name:         "lark 域名",
			input:        "https://company.larksuite.com/base/bascnABCDEF?table=tblXYZ123",
			wantAppToken: "bascnABCDEF",
			wantTableID:  "tblXYZ123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			appToken, tableID := ParseBitableURL(tc.input)
			if appToken != tc.wantAppToken {
				t.Errorf("appToken：got %q, want %q", appToken, tc.wantAppToken)
			}
			if tableID != tc.wantTableID {
				t.Errorf("tableID：got %q, want %q", tableID, tc.wantTableID)
			}
		})
	}
}
