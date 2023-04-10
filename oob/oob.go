package oob

import (
	"github.com/licairong/cloudboot-provider-framework/util"
	"strconv"
	"time"
)

const (
	// DefaultWorker 默认处理器名称
	DefaultWorker = "IPMI"
)

const (
	// DHCP IP来源-DHCP
	DHCP = "dhcp"
	// Static IP来源-静态IP
	Static = "static"
)

const (
	// PowerOn 设备电源状态-开机
	PowerOn = "on"
	// PowerOff 设备电源状态-关机
	PowerOff = "off"
)

// Worker OOB处理器
type Worker interface {
	NetworkWorker
	UserWorker
	BMCWorker
	// Name 返回处理器实现的名称
	Name() string
	// FRUDevice 返回物理机基本信息。
	// 若带外IP不可达，则返回IPUnreachableError错误。
	// 若用户名、密码不匹配，则返回UsernamePasswordError错误。
	FRUDevice() (*FRUDevice, error)
	// ValidateSN 校验预期的SN与实际的SN是否匹配
	ValidateSN(sn string) error
	// 返回电源状态
	PowerStatus() (status string, err error)
	// 设备上电开机
	PowerOn() error
	// 设备下电关机
	PowerOff() error
	// 设备重启
	PowerReset() error
	// 设备重启并指定其从网络引导
	PXEBoot(uefi bool, manufacturer string) error
	// Channel 返回impi channel
	Channel() (int, error)
	// PostCheck OOB配置实施后置检查
	PostCheck(sett *Setting) []*util.CheckingItem
	// Raw 发送原始的ipmi请求并返回响应内容
	Raw(args string) (response []byte, err error)
	// SelClear 日志清理
	SelClear() error
	// SensorList 返回物理机传感器信息。
	SensorList() ([]*SensorDevice, error)
	// SetSnmpTrap 设置snmptrap
	SetSnmpTrap(*SnmpSet) (err error)
}

// FRUDevice FRU设备基本信息
type FRUDevice struct {
	ProductManufacturer string // 物理机厂商名
	ProductName         string // 物理机产品名
	ProductSerial       string // 物理机序列号
}

// BMC OOB的BMC信息
type BMC struct {
	FirmwareReversion string
	IPMIVersion       string
	ManufacturerID    string
	ManufacturerName  string
}

// Access OOB channel access
type Access struct {
	Channel        int
	MaxUserIDs     int
	EnabledUserIDs int
	Accesses       []*UserAccess
}

// UserAccess OOB用户Access
type UserAccess struct {
	UserID             int
	UserName           string
	FixedName          string
	AccessAvailable    string
	LinkAuthentication string
	IPMIMessaging      string
	PrivilegeLevel     int
}

// Network OOB网络信息
type Network struct {
	IPSrc   string // 可选值: static|dhcp
	MAC     string // IP对应Mac地址
	IP      string // IP
	Netmask string // 子网掩码
	Gateway string // 默认网关
}

// NetworkWorker OOB网络模块处理器
type NetworkWorker interface {
	// SetDHCP 设置IP来源是DHCP
	SetDHCP() error
	// SetStaticIP 设置IP来源是静态IP
	SetStaticIP(ip, netmask, gateway string) error
	// Network 返回OOB网络信息
	Network() (*Network, error)
}

// User OOB用户信息
type User struct {
	Channel int
	ID      int
	Name    string
	Access  *UserAccess
}

// UserWorker OOB用户模块处理器
type UserWorker interface {
	// ChangeUserPassword 修改目标带外用户密码
	ChangeUserPassword(username, password string) (err error)
	// GenerateUser 生成用户帐号。
	// 若用户（以用户名为准）未存在，则新增帐号。
	// 若用户已经存在则修改用户密码、权限级别等属性。
	GenerateUser(sett *UserSettingItem) error
	// EnableUser 启用带外用户帐号
	EnableUser(username string) error
	// DisableUser 禁用带外用户帐号
	DisableUser(username string) error
	// Users 返回OOB用户列表
	Users() ([]*User, error)
}

// BMCWorker BMC模块处理器
type BMCWorker interface {
	// BMC 返回OOB的BMC信息
	BMC() (*BMC, error)
	// BMCColdReset (冷)重启BMC
	BMCColdReset() error
}

// Whoami 返回当前的BIOS固件对应的处理器名
func Whoami() (worker string, err error) {
	return DefaultWorker, nil // 暂时只有一个实现
}

const (
	// CriticalSeverity 日志严重性级别-严重
	CriticalSeverity = "critical"
	// WarningSeverity 日志严重性级别-警告
	WarningSeverity = "warning"
	// InformationSeverity 日志严重性级别-信息
	InformationSeverity = "information"
	// UnknownSeverity 日志严重性级别-未知
	UnknownSeverity = "unknown"
)

// EventLog 事件日志
type EventLog struct {
	SeqNumber string    `json:"seq_number"`
	Timestamp time.Time `json:"timestamp"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
}

// BySeqASC 为带外日志切片实现sort.Interface接口
type BySeqASC []*EventLog

func (a BySeqASC) Len() int      { return len(a) }
func (a BySeqASC) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySeqASC) Less(i, j int) bool {
	seqi, _ := strconv.Atoi(a[i].SeqNumber)
	seqj, _ := strconv.Atoi(a[j].SeqNumber)
	return seqi < seqj
}

// SensorDevice Sensor信息
type SensorDevice struct {
	Name     string //监控项
	Value    string //数值
	Units    string //单位
	State    string //状态
	Lonorec  string
	Locrit   string
	Lonocrit string
	Upcrit   string
	Upnocrit string
	Upnorec  string
}
type SnmpSet struct {
	Devicemodel           string
	SnmpTrapVersion       string
	SnmpV3User            string
	SnmpV3AuthPassword    string
	SnmpV3PrivPassword    string
	SnmpV3AuthProtocol    string
	SnmpV3PrivProtocol    string
	SnmpTrapAlarmseverity string
	CommunityName         string
	SnmpTrapEngineId      int
	SnmpTrapSystemName    string
	SnmpTrapSystemId      int
	SnmpTrapLocation      string
	SnmpTrapContact       string
	SnmpTrapHostOs        string
	SnmpTrapPortNo        int
	SnmpTrapServer        []*SnmpTrapServer
}

type SnmpTrapServer struct {
	TrapID              int
	SnmpTrapPolicy      string
	SnmpTrapChannel     int
	SnmpTrapType        string
	SnmpTrapDestination string
}
