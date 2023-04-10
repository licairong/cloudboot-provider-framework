package collector

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/licairong/cloudboot-provider-framework/util"
	strutil "github.com/licairong/cloudboot-provider-framework/util/strings"
)

var executor = util.NewBash()

// BASE 采集并返回当前设备的基本信息
func BASE() (*Base, error) {
	b := Base{
		SN:           string(readFileIgnoreErr("/sys/class/dmi/id/product_serial")),
		Manufacturer: string(readFileIgnoreErr("/sys/class/dmi/id/sys_vendor")),
		Model:        string(readFileIgnoreErr("/sys/class/dmi/id/product_name")),
	}

	if isVM, manufacturer, _ := vm(); isVM {
		b.IsVM = true
		b.Manufacturer = manufacturer
	}

	b.ChassisType, b.Height = chassis()
	b.Arch, _ = arch()
	return &b, nil
}

func vm() (yes bool, manufacturer string, err error) {
	output, err := executor.Exec(nil, "/bin/bash", "-c", `"dmesg | grep -i Hypervisor"`)
	if err != nil {
		return false, "", err
	}

	if strings.TrimSpace(string(output)) == "" {
		return false, "", nil
	}

	output = bytes.ToLower(output)
	if bytes.Contains(output, []byte("kvm")) {
		return true, "KVM", nil
	}
	if bytes.Contains(output, []byte("vmware")) {
		return true, "VMware", nil
	}
	return true, "unknown", nil
}

// readFileIgnoreErr 读取文件内容并忽略可能产生的错误
func readFileIgnoreErr(filename string) []byte {
	data, _ := ioutil.ReadFile(filename)
	return bytes.TrimSpace(data)
}

// chassis 返回设备类型及高度等信息
//
// # Below is the current list as of version 3.1.1 posted 1/13/2017.
// # https://www.dmtf.org/standards/smbios  main standards site
// # https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.1.1.pdf 2017 Standard doc
//
// # Other (1)
// # Unknown (2)
// # Desktop (3)
// # Low Profile Desktop (4)
// # Pizza Box (5)
// # Mini Tower (6)
// # Tower (7)
// # Portable (8)
// # Laptop (9)
// # Notebook (10)
// # Hand Held (11)
// # Docking Station (12)
// # All in One (13)
// # Sub Notebook (14)
// # Space-Saving (15)
// # Lunch Box (16)
// # Main System Chassis (17)
// # Expansion Chassis (18)
// # SubChassis (19)
// # Bus Expansion Chassis (20)
// # Peripheral Chassis (21)
// # RAID Chassis (22)
// # Rack Mount Chassis (23)
// # Sealed-case PC (24)
// # Multi-system chassis (25)
// # Compact PCI (26)
// # Advanced TCA (27)
// # Blade (28)
// # Blade Enclosure (29)
// # Tablet (30)
// # Convertible (31)
// # Detachable (32)
// # IoT Gateway (33)
// # Embedded PC (34)
// # Mini PC (35)
// # Stick PC (36)
func chassis() (chassisType string, height int) {
	chassisType = UnknownServer
	// dmidecode -t chassis
	if output, _ := executor.Exec(nil, "dmidecode", "-t", "chassis"); output != nil {
		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "Type:") {
				if strings.Contains(line, "Rack") {
					chassisType = RackServer
				}
			} else if strings.HasPrefix(line, "Height:") {
				sHeight := strings.TrimSpace(strings.TrimSuffix(strutil.ExtractValue(line, strutil.ColonSep), "U"))
				height, _ = strconv.Atoi(sHeight)
			}
		}
	}

	switch string(readFileIgnoreErr("/sys/class/dmi/id/chassis_type")) {
	case "23":
		chassisType = RackServer
	case "28":
		chassisType = BladeServer
	}
	return chassisType, height
}

// arch 返回设备的硬件架构
func arch() (string, error) {
	output, err := executor.Exec(nil, "uname", "-m")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
