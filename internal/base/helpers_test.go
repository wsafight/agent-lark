package base

import (
	"testing"
)

func TestParseFieldPair(t *testing.T) {
	tests := []struct {
		name    string
		pair    string
		wantKey string
		wantVal interface{}
		wantErr bool
	}{
		{
			name:    "普通字符串值",
			pair:    "姓名=张三",
			wantKey: "姓名",
			wantVal: "张三",
		},
		{
			name:    "整数值自动解析为 float64",
			pair:    "年龄=25",
			wantKey: "年龄",
			wantVal: float64(25),
		},
		{
			name:    "浮点数值",
			pair:    "分数=9.5",
			wantKey: "分数",
			wantVal: float64(9.5),
		},
		{
			name:    "布尔值 true",
			pair:    "已完成=true",
			wantKey: "已完成",
			wantVal: true,
		},
		{
			name:    "布尔值 false",
			pair:    "已删除=false",
			wantKey: "已删除",
			wantVal: false,
		},
		{
			name:    "null 值",
			pair:    "备注=null",
			wantKey: "备注",
			wantVal: nil,
		},
		{
			name:    "value 包含等号",
			pair:    "url=http://example.com/path=foo",
			wantKey: "url",
			wantVal: "http://example.com/path=foo",
		},
		{
			name:    "value 为空字符串",
			pair:    "备注=",
			wantKey: "备注",
			wantVal: "",
		},
		{
			name:    "value 为带引号的 JSON 字符串",
			pair:    `desc="hello world"`,
			wantKey: "desc",
			wantVal: "hello world",
		},
		{
			name:    "无等号，返回错误",
			pair:    "noequals",
			wantErr: true,
		},
		{
			name:    "key 为空，返回错误",
			pair:    "=value",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, val, err := parseFieldPair(tc.pair)
			if tc.wantErr {
				if err == nil {
					t.Errorf("期望错误，但得到 key=%q val=%v", key, val)
				}
				return
			}
			if err != nil {
				t.Fatalf("意外错误：%v", err)
			}
			if key != tc.wantKey {
				t.Errorf("key：got %q, want %q", key, tc.wantKey)
			}
			if val != tc.wantVal {
				t.Errorf("val：got %v (%T), want %v (%T)", val, val, tc.wantVal, tc.wantVal)
			}
		})
	}
}
