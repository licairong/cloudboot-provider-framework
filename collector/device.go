package collector

import (
	"fmt"
	byteutil "github.com/licairong/cloudboot-provider-framework/util/bytes"
)

// Device 设备信息采集上报请求数据
type Device struct {
	// 是否是虚拟机
	IsVM bool `json:"is_vm"`
	// 设备序列号
	SN string `json:"sn"`
	// 设备厂商名
	Manufacturer string `json:"manufacturer"`
	// 设备型号
	Model string `json:"model"`
	// 硬件架构
	Arch string `json:"arch"`
	// 设备类型
	ChassisType string `json:"chassis_type"`
	// 设备高度（U数）
	Height int `json:"height"`
	// 标识刀片在刀箱中的位置
	Locator int `json:"locator"`
	// 规格
	Spec string `json:"spec"`
	// CPU
	CPU *CPU `json:"cpu"`
	// 内存
	Memory *Memory `json:"memory"`
	// 主板
	Board *Motherboard `json:"board"`
	// 网卡
	NIC *NIC `json:"nic"`
	// 逻辑磁盘
	LogicalDisk *LogicalDisk `json:"logical_disk"`
	// 物理磁盘
	PhysicalDisk *PhysicalDisk `json:"physical_disk"`
	// RAID
	RAID *RAID `json:"raid"`
	// 带外
	OOB *OOB `json:"oob"`
	// BIOS
	BIOS *BIOS `json:"bios"`
	// 电源
	PowerSupply *PowerSupply `json:"power_supply"`
	// 风扇
	Fan *Fan `json:"fan"`
	// PCI插槽
	PCI *PCI `json:"pci"`
	// HBA
	HBA *HBA `json:"hba"`
	// LLDP
	LLDP *LLDP `json:"lldp"`
	// 自定义扩展字段
	Extra *Extra `json:"extra"`
	// BootOS网卡IP信息
	BootOSIP string `json:"bootos_ip"`
	// BootOS网卡Mac地址
	BootOSMAC string `json:"bootos_mac"`
}

const (
	// IPSourceStatic IP来源-static
	IPSourceStatic = "static"
	// IPSourceDHCP IP来源-dhcp
	IPSourceDHCP = "dhcp"
)

// Setup 设置
func (reqData *Device) Setup() {
	if reqData.OOB != nil && reqData.OOB.Network != nil {
		if reqData.OOB.Network.IPSrc == "Static Address" {
			reqData.OOB.Network.IPSrc = IPSourceStatic
		} else {
			reqData.OOB.Network.IPSrc = IPSourceDHCP
		}
	}

	if reqData.NIC != nil {
		for i := range reqData.NIC.Items {
			if reqData.NIC.Items[i].IP != "" {
				reqData.BootOSIP = reqData.NIC.Items[i].IP
				reqData.BootOSMAC = reqData.NIC.Items[i].MAC
			}
		}
	}

	// 计算规格
	var cores, memSizeGB, driveSizeGB int
	if reqData.CPU != nil && reqData.CPU.TotalCores > 0 {
		cores = reqData.CPU.TotalCores
	}
	if reqData.Memory != nil && reqData.Memory.TotalSize > 0 {
		memSizeGB = byteutil.Byte2GBRounding(byteutil.Byte(reqData.Memory.TotalSize))
	}
	if reqData.PhysicalDisk != nil && reqData.PhysicalDisk.TotalSize > 0 {
		driveSizeGB = byteutil.Byte2GBRounding(byteutil.Byte(reqData.PhysicalDisk.TotalSize))
	}
	reqData.Spec = fmt.Sprintf("%dC/%dG/%dG", cores, memSizeGB, driveSizeGB)
}

