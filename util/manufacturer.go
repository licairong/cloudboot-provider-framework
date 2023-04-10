package util

import (
	"github.com/licairong/cloudboot-provider-framework/util/collection"
	"strings"
)

const (
	// HP 厂商-惠普
	HP = "HP"
	// Dell 厂商-戴尔
	Dell = "Dell"
	// Huawei 厂商-华为
	Huawei = "Huawei"
	// Lenovo 厂商-联想
	Lenovo = "Lenovo"
	// H3C 厂商-华三
	H3C = "H3C"
	// Inspur 厂商-浪潮
	Inspur = "Inspur"
	// Sugon 厂商-曙光
	Sugon = "Sugon"
	// Supermicro 厂商-超微
	Supermicro = "Supermicro"
	// Greatwall 厂商-长城
	Greatwall = "Greatwall"
	// Gooxi 厂商-国鑫
	Gooxi = "Gooxi"
	// KVM 厂商-KVM虚拟机
	KVM = "KVM"
	// VMware 厂商-VMware虚拟机
	VMware = "VMware"
	// UNIS 厂商-紫光（收购H3C）
	UNIS = "UNIS"
	// Suma 厂商-曙光
	Suma = "Suma"
	// XFusion 厂商-超聚变
	XFusion = "xFusion"
	// RCSIT 厂商-天宫
	RCSIT = "RCSIT"
	// Suma 厂商-曙光
	ZTE = "ZTE"
	// IBM 厂商-IBM
	IBM = "IBM"
	// Baixin 厂商-百信
	Baixin = "Baixin"
	// DCN 厂商-神码
	DCN = "DCN"
	// Nettrix 厂商-宁畅
	Nettrix = "Nettrix"
)

// ValidManufacturer 返回是否是合法厂商名的布尔值
func ValidManufacturer(manufacturer string) bool {
	return collection.InSlice(manufacturer, []string{
		HP, Dell, Huawei, Lenovo, H3C, Inspur, Sugon, Supermicro, Greatwall, Gooxi, KVM, VMware, UNIS, Suma, XFusion, RCSIT, ZTE, IBM, Baixin, DCN, Nettrix,
	})
}

// ManufacturerName 返回统一的厂商命名
func ManufacturerName(manufacturer string) string {
	manufacturer = strings.TrimSpace(manufacturer)
	lowerInput := strings.ToLower(manufacturer)
	for i := range manufacturerMatchers {
		if manufacturerMatchers[i].Matcher(lowerInput) {
			return manufacturerMatchers[i].Manufacturer
		}
	}
	return manufacturer
}

var manufacturerMatchers = []struct {
	Manufacturer string
	Matcher      func(input string) bool
}{
	{
		Dell, func(input string) bool {
			return strings.Contains(input, "dell")
		},
	},
	{
		Supermicro, func(input string) bool {
			return strings.Contains(input, "super") && strings.Contains(input, "micro")
		},
	},
	{
		Greatwall, func(input string) bool {
			return strings.Contains(input, "great") && strings.Contains(input, "wall")
		},
	},
	{
		H3C, func(input string) bool {
			return strings.Contains(input, "h3c")
		},
	},
	{
		Lenovo, func(input string) bool {
			return strings.Contains(input, "lenovo") || strings.Contains(input, "ibm") || strings.Contains(input, "lnvo")
		},
	},
	{
		Huawei, func(input string) bool {
			return strings.Contains(input, "huawei")
		},
	},
	{
		Inspur, func(input string) bool {
			return strings.Contains(input, "inspur")
		},
	},
	{
		Sugon, func(input string) bool {
			return strings.Contains(input, "sugon")
		},
	},
	{
		Gooxi, func(input string) bool {
			return strings.Contains(input, "gooxi")
		},
	},
	{
		HP, func(input string) bool {
			return strings.Contains(input, "hp") ||
				strings.Contains(input, "hpe") ||
				(strings.Contains(input, "hewlett") && strings.Contains(input, "packard"))
		},
	},
	{
		KVM, func(input string) bool {
			return strings.Contains(input, "kvm")
		},
	},
	{
		VMware, func(input string) bool {
			return strings.Contains(input, "vmware")
		},
	},
	{
		UNIS, func(input string) bool {
			return strings.Contains(input, "unis")
		},
	},
	{
		Suma, func(input string) bool {
			return strings.Contains(input, "suma")
		},
	},
	{
		XFusion, func(input string) bool {
			return strings.Contains(input, "xfusion")
		},
	},
	{
		RCSIT, func(input string) bool {
			return strings.Contains(input, "rcsit")
		},
	},
	{
		ZTE, func(input string) bool {
			return strings.Contains(input, "zte")
		},
	},
	{
		IBM, func(input string) bool {
		return strings.Contains(input, "ibm")
		},
	},
	{
		Baixin, func(input string) bool {
		return strings.Contains(input, "baixin")
		},
	},
	{
		DCN, func(input string) bool {
		return strings.Contains(input, "dcn")
		},
	},
	{
		Nettrix, func(input string) bool {
		return strings.Contains(input, "nettrix")
		},
	},
}
