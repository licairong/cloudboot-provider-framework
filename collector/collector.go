package collector

import (
	"errors"
	"github.com/licairong/cloudboot-provider-framework/oob"
	"github.com/licairong/cloudboot-provider-framework/util"
	"reflect"

	jsoniter "github.com/json-iterator/go"

	//"idcos.io/cloudboot/hardware/v4"
	//"idcos.io/cloudboot/hardware/v4/oob"
	//"idcos.io/cloudboot/hardware/v4/raid"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	// DefaultCollector 默认采集器名称
	DefaultCollector = "BOOTOS"
)

var (
	// ErrUnregisteredCollector 未注册的采集器
	ErrUnregisteredCollector = errors.New("unregistered collector")
	// ErrNotSupported 暂不支持
	ErrNotSupported = errors.New("not supported yet")
)

// Collector 设备信息采集器
type Collector interface {
	// Destroy 资源回收并销毁采集器
	Destroy() error
	// SetDebug 设置是否开启debug。若开启debug，会将关键日志信息写入console。
	// SetDebug(debug bool)
	// SetLog 更换日志实现。默认情况下内部无日志实现。
	SetLog(log util.Logger)
	// 采集并返回当前设备基本信息
	BASE() (*Base, error)
	// CPU 采集并返回当前设备的CPU信息
	CPU() (*CPU, error)
	// Memory 采集并返回当前设备内存信息
	Memory() (*Memory, error)
	// Motherboard 采集并返回当前设备的主板信息
	Motherboard() (*Motherboard, error)
	// LogicalDisk 采集并返回当前设备的逻辑磁盘信息
	LogicalDisk() (*LogicalDisk, error)
	// PhysicalDisk 采集并返回当前设备的物理磁盘信息
	PhysicalDisk(uri, user_name, password string) (*PhysicalDisk, error)
	// NIC 采集并返回当前设备的网卡信息
	NIC() (*NIC, error)
	// HBA 采集并返回当前设备的HBA卡信息。若当前设备不包含HBA卡，则返回值都为nil。
	HBA() (*HBA, error)
	// OOB 采集并返回当前设备的OOB信息。
	OOB() (*OOB, error)
	// BIOS 采集并返回当前设备的BIOS信息。
	BIOS() (*BIOS, error)
	// RAID 采集并返回当前设备的RAID信息。
	RAID(uri, user_name, password string) (*RAID, error)
	// PCI 采集并返回当前设备的所有PCI插槽信息。
	PCI() (*PCI, error)
	// Fan 采集并返回当前设备的所有风扇信息。
	Fan() (*Fan, error)
	// PowerSupply 采集并返回当前设备的电源信息。
	PowerSupply() (*PowerSupply, error)
	// LLDP 采集并返回当前设备LLDP信息。
	LLDP() (*LLDP, error)
	// IDRAC 采集并返回Dell iDRAC信息。
	IDRAC() (*IDRAC, error)
	// ILO 采集并返回HP iLO信息。
	ILO() (*ILO, error)
	// Backplane 采集并返回背板信息。
	Backplane() (*Backplane, error)
	// EventLogs 采集并返回带外事件日志
	EventLogs() ([]*oob.EventLog, error)
	// Extra 执行采集脚本并返回采集到的信息。若执行采集脚本时发生错误，则丢弃该错误，继续执行后续脚本。
	// 采集脚本需满足以下条件：
	// 1、脚本的开始位置通过shebang指定执行该脚本的程序，如'#! /usr/bin/env python'。
	// 2、脚本执行完毕后，需要将采集到的设备信息以JSON Object格式字符串写入stdout。
	// 3、多个脚本输出的JSON Object属性名不能重复，否则在合并多个JSON Object过程中可能导致数据丢失。
	Extra(scripts [][]byte) *Extra
	//Snmp校验
	Check() error
	//退出redfish
	LogoutRedfish()
}

const (
	// UnknownServer 未知类型服务器
	UnknownServer = "unknown"
	// RackServer 机架式服务器
	RackServer = "rack_server"
	// BladeServer 刀片式服务器
	BladeServer = "blade_server"
	// Minicomputer 小型机
	Minicomputer = "minicomputer"
)

// Base 基本信息
type Base struct {
	IsVM         bool   `json:"is_vm"`        // 是否是虚拟机
	SN           string `json:"sn"`           // 序列号
	Manufacturer string `json:"manufacturer"` // 厂商
	Model        string `json:"model"`        // 型号
	Arch         string `json:"arch"`         // 硬件架构
	ChassisType  string `json:"chassis_type"` // 类型。可选值：rack_server|blade_server|minicomputer|unknown
	Height       int    `json:"height"`       // 高度（U数）
}

// Backplane 背板
type Backplane struct {
	Items []*BackplaneItem `json:"items"`
}

// BackplaneItem 背板条目
type BackplaneItem struct {
	Name            string `json:"name"`             // 背板名称
	FirmwareVersion string `json:"firmware_version"` // 背板固件版本号
}

// ILO HP iLO
type ILO struct {
	FirmwareDate    string `json:"firmware_date"`
	FirmwareVersion string `json:"firmware_version"`
}

// IDRAC Dell iDRAC
type IDRAC struct {
	RACDateTime        string `json:"rac_date_time"`
	FirmwareVersion    string `json:"firmware_version"`
	FirmwareBuild      string `json:"firmware_build"`
	LastFirmwareUpdate string `json:"last_firmware_update"`
}

// NIC 网卡
type NIC struct {
	Items []*NICPort `json:"items"`
}

// NICPort 网口
type NICPort struct {
	Location        string     `json:"location" comment:"位置"`           // 硬件位置/名称
	MAC             string     `json:"mac" comment:"MAC"`               // mac地址
	IP              string     `json:"ip"`                              // IP（非硬件层面属性）
	Port            int        `json:"port" comment:"网口"`               // 网口
	PCISlot         int        `json:"pci_slot" comment:"PCI槽位"`        // PCI槽位。若为0，则表示板载网卡。
	BusAddress      string     `json:"bus_address" comment:"总线地址"`      // 总线地址
	Speed           string     `json:"speed" comment:"速率"`              // 速率
	Manufacturer    string     `json:"manufacturer" comment:"厂商"`       // 厂商
	Model           string     `json:"model" comment:"型号"`              // 型号
	FirmwareVersion string     `json:"firmware_version" comment:"固件版本"` // 固件版本
	Side            string     `json:"side" comment:"内/外置"`             // 内/外置
	Type            string     `json:"type"`                            // 类型。可选值：Ethernet
	Link            string     `json:"link"`                            // 网卡链接状态。可选值：yes|no
	SwitchRef       *SwitchRef `json:"switch_ref"`                      // 关联交换机信息
}

// SwitchRef 关联交换机信息，用于nic和设备
type SwitchRef struct {
	Name          string `json:"name"`
	Mac           string `json:"mac"`
	NICMac        string `json:"nic_mac"`
	BridgeEnabled string `json:"bridge_enabled"`
	RouteEnabled  string `json:"route_enabled"`
	MgmtIP        string `json:"mgmt_ip"`     // 管理IP
	Descr         string `json:"descr"`       // 描述信息
	TTL           string `json:"ttol"`        // 超时时间
	PortIfName    string `json:"port_ifname"` // 端口
	PortMFS       string `json:"port_mfs"`
	Vlan          string `json:"vlan"`
	PPVIDEnabled  string `json:"ppvid_enabled"`
	PPVIDSupport  string `json:"ppvid_support"`
}

// Equal 返回是否相同的布尔值
func (m *NICPort) Equal(t *NICPort) bool {
	if m == nil || t == nil {
		return false
	}
	if m == t {
		return true
	}
	return m.Location == t.Location &&
		m.MAC == t.MAC &&
		m.Port == t.Port &&
		m.PCISlot == t.PCISlot &&
		m.BusAddress == t.BusAddress &&
		m.Speed == t.Speed &&
		m.Manufacturer == t.Manufacturer &&
		m.Model == t.Model &&
		m.FirmwareVersion == t.FirmwareVersion &&
		m.Type == t.Type
}

// ToJSON 序列化为JSON
func (nic NIC) ToJSON() []byte {
	b, _ := json.Marshal(nic)
	return b
}

// CPU CPU信息。
// 总核数 = 物理CPU个数 X 每颗物理CPU的核数
type CPU struct {
	TotalThreads   int          `json:"total_threads"`   // 总超线程数
	TotalCores     int          `json:"total_cores"`     // 物理CPU总核心数
	TotalPhysicals int          `json:"total_physicals"` // 物理CPU数量
	Items          []*Processor `json:"items"`           // 物理CPU列表
}

// Processor 物理CPU
type Processor struct {
	SocketDesignation string   `json:"socket_designation"` // 槽位
	Manufacturer      string   `json:"manufacturer"`       // 厂商
	Family            string   `json:"family"`             // 系列
	Model             string   `json:"model"`              // 型号
	Type              string   `json:"type"`               // 类型
	MaxSpeed          string   `json:"max_speed"`          // 最大主频
	CurrentSpeed      string   `json:"current_speed"`      // 当前主频
	Cores             int      `json:"cores"`              // 单颗物理CPU的核心数
	EnabledCores      int      `json:"enabled_cores"`      // 启用核心数
	Threads           int      `json:"threads"`            // 超线程数
	Voltage           string   `json:"voltage"`            // 电压
	Flags             []string `json:"flags"`              // 指令集
	L1Cache           string   `json:"l1_cache"`           // 一级缓存
	L2Cache           string   `json:"l2_cache"`           // 二级缓存
	L3Cache           string   `json:"l3_cache"`           // 三级缓存
}

// ToJSON 序列化为JSON
func (cpu CPU) ToJSON() []byte {
	b, _ := json.Marshal(cpu)
	return b
}

// Memory 内存
type Memory struct {
	NumberOfDevices     int             `json:"number_of_devices"`      // 内存插槽数量
	NumberOfUsedDevices int             `json:"number_of_used_devices"` // 已用内存插槽数量
	MaximumSize         int64           `json:"maximum_size"`           // 能识别到的最大容量（单位Byte）
	TotalSize           int64           `json:"total_size"`             // 当前内存总容量（单位Byte）
	Items               []*MemoryDevice `json:"items"`                  // 物理内存条列表
}

// MemoryDevice 插在内存插槽上的内存条
type MemoryDevice struct {
	Location          string `json:"location" comment:"位置"`       // 硬件位置
	Size              int64  `json:"size" comment:"容量"`           // 容量（单位Byte）
	Type              string `json:"type" comment:"类型"`           // 类型，如DDR3
	Speed             string `json:"speed" comment:"速率"`          // 速率
	Manufacturer      string `json:"manufacturer" comment:"厂商"`   // 厂商
	SerialNumber      string `json:"serial_number" comment:"序列号"` // 序列号
	PartNumber        string `json:"part_number" comment:"部件号"`   // 部件号
	AssetTag          string `json:"asset_tag"`                   // 资产标签
	ConfiguredVoltage string `json:"configured_voltage"`          // 配置电压（单位V）
}

// Equal 返回是否相同的布尔值
func (m *MemoryDevice) Equal(t *MemoryDevice) bool {
	if m == nil || t == nil {
		return false
	}
	if m == t {
		return true
	}
	return m.Location == t.Location &&
		m.Size == t.Size &&
		m.Type == t.Type &&
		m.Speed == t.Speed &&
		m.Manufacturer == t.Manufacturer &&
		m.SerialNumber == t.SerialNumber &&
		m.PartNumber == t.PartNumber
}

// ToJSON 序列化为JSON
func (m Memory) ToJSON() []byte {
	b, _ := json.Marshal(m)
	return b
}

// LogicalDisk 逻辑磁盘
type LogicalDisk struct {
	TotalSize int64           `json:"total_size"` // 磁盘总容量，单位Byte。
	Items     []*LogicalDrive `json:"items"`
}

// LogicalDrive 逻辑驱动器条目
type LogicalDrive struct {
	Name       string `json:"name"`         // 名称，如'/dev/sda'
	ByPathName string `json:"by_path_name"` // by-path持久化名称，如'pci-0000:03:00.0-scsi-0:0:0:0'
	Size       string `json:"size"`         // 容量，如'599.6 GB, 599550590976 bytes'
}

// ToJSON 序列化为JSON
func (d LogicalDisk) ToJSON() []byte {
	b, _ := json.Marshal(d)
	return b
}

// PhysicalDisk 物理驱动器（物理硬盘）
type PhysicalDisk struct {
	TotalSize int64            `json:"total_size"` // 总容量，单位Byte。
	Items     []*PhysicalDrive `json:"items"`      // 物理硬盘列表
}

// PhysicalDrive 物理驱动器条目
type PhysicalDrive struct {
	Location        string `json:"location" comment:"位置"`             // 硬件位置/名称
	Slot            string `json:"slot" comment:"插槽编号"`               // 硬盘插槽编号
	Manufacturer    string `json:"manufacturer" comment:"厂商"`         // 厂商
	Model           string `json:"model" comment:"型号"`                // 型号
	WWN             string `json:"wwn" comment:"WWN"`                 // wwn
	SerialNumber    string `json:"serial_number" comment:"序列号"`       // 序列号
	BusType         string `json:"bus_type" comment:"总线类型"`           // 总线类型
	MediaType       string `json:"media_type" comment:"媒体类型"`         // 媒体类型
	Size            int64  `json:"size" comment:"容量"`                 // 容量，单位byte。
	PartNumber      string `json:"part_number" comment:"部件号"`         // 部件号
	FirmwareVersion string `json:"firmware_version" comment:"固件版本"`   // 固件版本
	ErrorCount      int    `json:"error_count" comment:"错误数"`         // 错误数
	TransferSpeed   string `json:"transfer_speed" comment:"传输速率"`     // 传输速率
	FirmwareState   string `json:"firmware_state" comment:"固件状态"`     // 固件状态
	ForeignState    string `json:"foreign_state" comment:"掉线状态"`      // 掉线状态
	InquiryData     string `json:"inquiry_data" comment:"诊断数据"`       // 诊断数据
	ControllerID    string `json:"controller_id" comment:"RAID控制器ID"` // RAID控制器ID
}

// Equal 返回是否相同的布尔值
func (m *PhysicalDrive) Equal(t *PhysicalDrive) bool {
	if m == nil || t == nil {
		return false
	}
	if m == t {
		return true
	}
	return reflect.DeepEqual(m, t)
}

// ToJSON 序列化为JSON
func (d PhysicalDisk) ToJSON() []byte {
	b, _ := json.Marshal(d)
	return b
}

// Motherboards 主板
type Motherboards struct {
	Items []Motherboard `json:"motherboard"`
}

// Motherboard 主板
type Motherboard struct {
	Manufacturer    string           `json:"manufacturer"`     // 厂商名
	ProductName     string           `json:"product_name"`     // 产品名
	SerialNumber    string           `json:"serial_number"`    // 序列号
	FirmwareVersion string           `json:"firmware_version"` // 固件版本号
	OnboardDevices  []*OnboardDevice `json:"onboard_devices"`  // 主板上的部件
}

// OnboardDevice 主板上的部件信息
type OnboardDevice struct {
	ReferenceDesignation string `json:"reference_designation"` // 部件
	Type                 string `json:"type"`                  // 类型
	Status               string `json:"status"`                // 状态
	BusAddress           string `json:"bus_address"`           // 总线地址
}

// ToJSON 序列化为JSON
func (board Motherboard) ToJSON() []byte {
	b, _ := json.Marshal(board)
	return b
}

// RAID RAID
type RAID struct {
	Items []*RaidController `json:"items"`
}

// RAIDArray RAID阵列
type RAIDArray struct {
	Level string `json:"level"`
}

// Controller RAID控制器
type RaidController struct {
	ID              string `json:"id" comment:"编号"` // 唯一标识/编号
	Manufacturer    string `json:"manufacturer" comment:"厂商"`
	Model           string `json:"model" comment:"型号"`              // 型号
	FirmwareVersion string `json:"firmware_version" comment:"固件版本"` // 固件版本
	SerialNumber    string `json:"serial_number" comment:"序列号"`     // 序列号
	PCIAddress      string `json:"pci_address" comment:"PCI总线地址"`   // PCI地址
	Mode            string `json:"mode"`                            // 当前所处模式RAID/JBOD(HBA)
}

// Equal 返回是否相同的布尔值
func (m *RaidController) Equal(t *RaidController) bool {
	if m == nil || t == nil {
		return false
	}
	if m == t {
		return true
	}
	return m.ID == t.ID &&
		m.Model == t.Model &&
		m.FirmwareVersion == t.FirmwareVersion &&
		m.SerialNumber == t.SerialNumber &&
		m.PCIAddress == t.PCIAddress
}

// ToJSON 序列化为JSON
func (raid RAID) ToJSON() []byte {
	b, _ := json.Marshal(raid)
	return b
}

// OOB OOB
type OOB struct {
	Network         *OOBNetwork `json:"network"`  // 网络
	User            []*OOBUser  `json:"user"`     // 用户
	FirmwareVersion string      `json:"firmware"` // 固件版本
}

// OOBNetwork 带外网络信息
type OOBNetwork struct {
	IPSrc   string `json:"ip_src"`  // IP来源。可选值: static|dhcp
	IP      string `json:"ip"`      // IP地址
	MAC     string `json:"mac"`     // 硬件地址
	Netmask string `json:"netmask"` // 子网掩码
	Gateway string `json:"gateway"` // 默认网关
}

// OOBUser 带外用户信息
type OOBUser struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	PrivilegeLevel int    `json:"privilege_level"`
}

// ToJSON 序列化为JSON
func (oob OOB) ToJSON() []byte {
	b, _ := json.Marshal(oob)
	return b
}

// BIOS BIOS
type BIOS struct {
	Manufacturer    string   `json:"manufacturer" comment:"厂商"`       // 厂商
	FirmwareVersion string   `json:"firmware_version" comment:"固件版本"` // 固件版本号
	ReleaseDate     string   `json:"release_date" comment:"发布时间"`     // 发布时间
	Characteristics []string `json:"characteristics"`                 // 特性
}

// Equal 返回是否相同的布尔值
func (m *BIOS) Equal(t *BIOS) bool {
	if m == nil || t == nil {
		return false
	}
	if m == t {
		return true
	}
	return m.Manufacturer == t.Manufacturer &&
		m.FirmwareVersion == t.FirmwareVersion &&
		m.ReleaseDate == t.ReleaseDate
}

// ToJSON 序列化为JSON
func (m BIOS) ToJSON() []byte {
	b, _ := json.Marshal(m)
	return b
}

// PowerSupply 电源
type PowerSupply struct {
	Items []*SystemPowerSupply `json:"items"`
}

// SystemPowerSupply 电源
type SystemPowerSupply struct {
	Location         string `json:"location" comment:"位置"`              // 位置
	Name             string `json:"name"`                               // 名称
	Manufacturer     string `json:"manufacturer" comment:"厂商"`          // 厂商
	SerialNumber     string `json:"serial_number" comment:"序列号"`        // 序列号
	PartNumber       string `json:"part_number" comment:"部件号"`          // 部件号
	AssetTag         string `json:"asset_tag"`                          // 资产标签
	FirmwareVersion  string `json:"firmware_version" comment:"固件版本"`    // 固件版本号
	Model            string `json:"model" comment:"型号"`                 // 型号
	InputVoltage     string `json:"input_voltage" comment:"输入电压"`       // 输入电压
	TotalOutputPower string `json:"total_output_power" comment:"总输出功率"` // 总输出功率
	Plugged          string `json:"plugged"`                            // 是否可插拔
	HotReplaceable   string `json:"hot_replaceable"`                    // 是否支持热插拔
}

// Equal 返回是否相同的布尔值
func (m *SystemPowerSupply) Equal(t *SystemPowerSupply) bool {
	if m == nil || t == nil {
		return false
	}
	if m == t {
		return true
	}
	return m.Location == t.Location &&
		m.Manufacturer == t.Manufacturer &&
		m.SerialNumber == t.SerialNumber &&
		m.PartNumber == t.PartNumber &&
		m.FirmwareVersion == t.FirmwareVersion &&
		m.Model == t.Model &&
		m.InputVoltage == t.InputVoltage &&
		m.TotalOutputPower == t.TotalOutputPower
}

// ToJSON 序列化为JSON
func (p PowerSupply) ToJSON() []byte {
	b, _ := json.Marshal(p)
	return b
}

// Fan 风扇
type Fan struct {
	Items []*FanItem `json:"items"`
}

// FanItem 风扇
type FanItem struct {
	Location string `json:"location"` // 位置/名称
	Speed    string `json:"speed"`    // 转速
}

// ToJSON 序列化为JSON
func (fan Fan) ToJSON() []byte {
	b, _ := json.Marshal(fan)
	return b
}

// PCI PCI
type PCI struct {
	TotalSlots int           `json:"total_slots"`
	Items      []*SystemSlot `json:"items"`
}

// SystemSlot PCI插槽
type SystemSlot struct {
	ID           int        `json:"id"`            // 槽位号
	Designation  string     `json:"designation"`   // 槽位名
	Type         string     `json:"type"`          // 设备类型
	CurrentUsage string     `json:"current_usage"` // 当前使用情况
	Length       string     `json:"length"`        // 长度
	BusAddress   string     `json:"bus_address"`   // 总线地址
	PCIDevice    *PCIDevice `json:"pci_device"`    // PCI设备
}

// PCIDevice PCI设备
type PCIDevice struct {
	Name            string `json:"name"`             // 名称
	Type            string `json:"type"`             // 类型
	Manufacturer    string `json:"manufacturer"`     // 厂商
	Model           string `json:"model"`            // 型号
	Speed           string `json:"speed"`            // 速率
	FirmwareVersion string `json:"firmware_version"` // 固件版本号
}

// ToJSON 序列化为JSON
func (pci PCI) ToJSON() []byte {
	b, _ := json.Marshal(pci)
	return b
}

// HBA HBA卡
type HBA struct {
	Items []*HBAPort `json:"items"`
}

// HBAPort HBA端口
type HBAPort struct {
	Host            string `json:"host"`             // 主机标识
	State           string `json:"state"`            // port的状态。可选值：Offline|Online
	WWPN            string `json:"wwpn"`             // World Wide Port Name（port自身的WWN）
	WWNN            string `json:"wwnn"`             // World Wide Node Name（port所属HBA卡的WWN）
	FirmwareVersion string `json:"firmware_version"` // 固件版本号
}

// ToJSON 序列化为JSON
func (hba HBA) ToJSON() []byte {
	b, _ := json.Marshal(hba)
	return b
}

// LLDP LLDP采集到的交换机信息
type LLDP map[string]interface{}

// ToJSON 序列化为JSON
func (lldp LLDP) ToJSON() []byte {
	if lldp == nil {
		return nil
	}
	b, _ := json.Marshal(lldp)
	return b
}

// Extra 自定义扩展属性
type Extra map[string]interface{}

// ToJSON 序列化为JSON
func (extra Extra) ToJSON() []byte {
	b, _ := json.Marshal(extra)
	return b
}
