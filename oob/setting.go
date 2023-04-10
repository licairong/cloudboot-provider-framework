package oob

import "strings"

const (
	// NoAccessLevel 带外用户权限级别-未知
	NoAccessLevel = 0
	// CallbackLevel 带外用户权限级别-callback
	CallbackLevel = 1
	// UserLevel 带外用户权限级别-user
	UserLevel = 2
	// OperatorLevel 带外用户权限级别-operator
	OperatorLevel = 3
	// AdministratorLevel 带外用户权限级别-administrator
	AdministratorLevel = 4
	// OEMProprietaryLevel 带外用户权限级别-OEM
	OEMProprietaryLevel = 5
)

// StringUserLevel 返回用户级别的字符串标识符
func StringUserLevel(l int) string {
	switch l {
	case NoAccessLevel:
		return "NoAccessLevel"
	case CallbackLevel:
		return "CallbackLevel"
	case UserLevel:
		return "UserLevel"
	case OperatorLevel:
		return "OperatorLevel"
	case AdministratorLevel:
		return "AdministratorLevel"
	case OEMProprietaryLevel:
		return "OEMProprietaryLevel"
	}
	panic("unsupported level")
}

// IntUserLevel 根据字符串返回整数类型的权限级别
func IntUserLevel(s string) int {
	s = strings.ToLower(s)
	if strings.HasPrefix(s, "no access") {
		return NoAccessLevel
	} else if strings.HasPrefix(s, "callback") {
		return CallbackLevel
	} else if strings.HasPrefix(s, "user") {
		return UserLevel
	} else if strings.HasPrefix(s, "operator") {
		return OperatorLevel
	} else if strings.HasPrefix(s, "admin") {
		return AdministratorLevel
	} else if strings.HasPrefix(s, "oem") {
		return OEMProprietaryLevel
	}
	return -1
}

// Setting OOB配置参数
type Setting struct {
	Network *NetworkSetting `json:"network"`
	User    *UserSetting    `json:"user"`
	BMC     *BMCSetting     `json:"bmc"`
}

// NetworkSetting OOB配置参数-网络
type NetworkSetting struct {
	IPSrc    string `json:"ip_src"` // IP来源。可选值: static|dhcp
	StaticIP struct {
		IP      string `json:"ip,omitempty"`
		Netmask string `json:"netmask,omitempty"`
		Gateway string `json:"gateway,omitempty"`
	} `json:"static_ip,omitempty"`
}

const (
	// EnabledUser 用户帐号状态-启动
	EnabledUser = "enabled"
	// DisabledUser 用户帐号状态-禁用
	DisabledUser = "disabled"
)

// UserSetting OOB配置参数-用户
type UserSetting []*UserSettingItem

// UserSettingItem 带外用户信息
type UserSettingItem struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	PrivilegeLevel int    `json:"privilege_level"`
	Status         string `json:"status"` // 帐号状态。可选值: enabled|disabled
}

// BMCSetting OOB配置参数-电源
type BMCSetting struct {
	ColdRest string `json:"cold_reset,omitempty"` // BMC冷启动。可选值: ON|OFF
}
