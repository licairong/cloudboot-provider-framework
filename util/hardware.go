package util

import (
	"strings"
)

const (
	// MatchedYES 实际值与预期值相符
	MatchedYES = "yes"
	// MatchedNO 实际值与预期值不符
	MatchedNO = "no"
	// MatchedUnknown 实际值与预期值是否相符未知
	MatchedUnknown = "unknown"
)

// CheckingItem 硬件配置实施后置检查结果
type CheckingItem struct {
	Title    string `json:"title"`    // 检查项名称
	Expected string `json:"expected"` // 预期值
	Actual   string `json:"actual"`   // 实际值
	Matched  string `json:"matched"`  // 是否匹配
	Error    string `json:"error"`    // 检查过程出错信息
}

// CheckingResult 硬件配置实施后置检查结果集合
type CheckingResult struct {
	RAIDItems []*CheckingItem `json:"raid"`
	OOBItems  []*CheckingItem `json:"oob"`
	BIOSItems []*CheckingItem `json:"bios"`
	FWItems   []*CheckingItem `json:"firmware"`
	//...继续扩展
}

// MatchFunc 对比预期值与实际值是否匹配的匹配器
type MatchFunc func(expected, actual string) bool

// EqualMatch 相等匹配器
func EqualMatch(expected, actual string) bool {
	return expected == actual
}

// EqualIgnoreCaseMatch 大小写不敏感相等匹配器
func EqualIgnoreCaseMatch(expected, actual string) bool {
	return strings.ToLower(expected) == strings.ToLower(actual)
}

// ContainsMatch 内容包含匹配器
func ContainsMatch(expected, actual string) bool {
	return strings.Contains(actual, expected)
}

// CheckingHelper 硬件配置实施检查帮手。
// 检查帮手默认使用相等匹配模式。
type CheckingHelper struct {
	item    *CheckingItem
	matcher MatchFunc
}

// NewCheckingHelper 实例化检查帮手。默认使用相等匹配器。
func NewCheckingHelper(title, expected, actual string) *CheckingHelper {
	return &CheckingHelper{
		item: &CheckingItem{
			Title:    title,
			Expected: expected,
			Actual:   actual,
		},
		matcher: EqualMatch,
	}
}

// Matcher 重置匹配器
func (h *CheckingHelper) Matcher(matcher MatchFunc) *CheckingHelper {
	h.matcher = matcher
	return h
}

// Do 执行检查并返回检查项结果。
func (h *CheckingHelper) Do() *CheckingItem {
	if h.matcher(h.item.Expected, h.item.Actual) {
		h.item.Matched = MatchedYES
	} else {
		h.item.Matched = MatchedNO
	}
	return h.item
}

// IsInvalidOptionError 判断是否是无效命令行选项错误，若是返回true，反之为false。
func IsInvalidOptionError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*InvalidOptionError)
	return ok
}

// InvalidOptionError 无效的命令行选项错误
type InvalidOptionError struct {
	Name  string // 选项名
	Value string // 选项值
	// Err   error  // 错误信息
}

// NewInvalidOptionError 返回无效的命令行选项错误实例
func NewInvalidOptionError(fields ...string) error {
	var name, value string
	switch len(fields) {
	case 0:
	case 1:
		name = fields[0]
	default:
		name = fields[0]
		value = fields[1]
	}
	return &InvalidOptionError{
		Name:  name,
		Value: value,
	}
}

func (e *InvalidOptionError) Error() string {
	var buf strings.Builder
	buf.WriteString("invalid option error")
	if e.Name != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Name)
	}
	if e.Value != "" {
		if e.Name != "" {
			buf.WriteString("=")
		} else {
			buf.WriteString(": ")
		}
		buf.WriteString(e.Value)
	}
	return buf.String()
}
