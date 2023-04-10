package strings

import "strings"

const (
	// EqSep 等号分隔符
	EqSep = "="
	// ColonSep 冒号分隔符
	ColonSep = ":"
)

// ExtractValue 截取kv对中v的内容。若kv内容为"name : voidint"，那么将返回"voidint"。
func ExtractValue(kv, sep string) (value string) {
	if !strings.Contains(kv, sep) {
		return kv
	}
	return strings.TrimSpace(strings.SplitN(kv, sep, 2)[1])
}
