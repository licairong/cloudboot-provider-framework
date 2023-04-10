package ipmi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"bou.ke/monkey"
	. "github.com/smartystreets/goconvey/convey"

	"idcos.io/cloudboot/hardware/v4"
	"idcos.io/cloudboot/hardware/v4/oob"
)

func TestPowerStatus(t *testing.T) {
	Convey("查询设备电源状态", t, func() {
		Convey("命令执行失败", func() {
			var ErrExec = errors.New("Unable to establish IPMI v2 / RMCP+ session")
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				return nil, ErrExec
			})
			defer monkey.UnpatchAll()

			status, err := NewWorker().PowerStatus()
			So(err, ShouldEqual, ErrExec)
			So(status, ShouldBeBlank)
		})

		Convey("已开机", func() {
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				So(cmd, ShouldEqual, tool)
				So(len(args), ShouldEqual, 3)
				So(strings.TrimSpace(strings.Join(args, " ")), ShouldEqual, "power status")
				return ioutil.ReadFile("testdata/ipmitool_power_status_on.txt")
			})
			defer monkey.UnpatchAll()

			status, err := NewWorker().PowerStatus()
			So(err, ShouldBeNil)
			So(status, ShouldEqual, oob.PowerOn)
		})

		Convey("已关机", func() {
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				So(cmd, ShouldEqual, tool)
				So(len(args), ShouldEqual, 3)
				So(strings.TrimSpace(strings.Join(args, " ")), ShouldEqual, "power status")
				return ioutil.ReadFile("testdata/ipmitool_power_status_off.txt")
			})
			defer monkey.UnpatchAll()

			status, err := NewWorker().PowerStatus()
			So(err, ShouldBeNil)
			So(status, ShouldEqual, oob.PowerOff)
		})
	})
}

func TestPowerOn(t *testing.T) {
	Convey("开机", t, func() {
		Convey("当前已开机，无需操作。", func() {
			var w *worker
			monkey.PatchInstanceMethod(reflect.TypeOf(w), "PowerStatus", func(b *worker) (string, error) {
				return oob.PowerOn, nil
			})
			defer monkey.UnpatchAll()

			So(NewWorker().PowerOn(), ShouldBeNil)
		})

		Convey("当前已关机，执行开机", func() {
			var w *worker
			monkey.PatchInstanceMethod(reflect.TypeOf(w), "PowerStatus", func(b *worker) (string, error) {
				return oob.PowerOff, nil
			})

			var ErrPowerOn = errors.New("power on error")
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				So(cmd, ShouldEqual, tool)
				So(len(args), ShouldEqual, 3)
				So(strings.TrimSpace(strings.Join(args, " ")), ShouldEqual, "power on")
				return nil, ErrPowerOn
			})

			defer monkey.UnpatchAll()

			So(NewWorker().PowerOn(), ShouldEqual, ErrPowerOn)
		})
	})
}

func TestPowerOff(t *testing.T) {
	Convey("关机", t, func() {
		Convey("当前已关机，无需操作。", func() {
			var w *worker
			monkey.PatchInstanceMethod(reflect.TypeOf(w), "PowerStatus", func(b *worker) (string, error) {
				return oob.PowerOff, nil
			})
			defer monkey.UnpatchAll()

			So(NewWorker().PowerOff(), ShouldBeNil)
		})

		Convey("当前已开机，执行关机", func() {
			var w *worker
			monkey.PatchInstanceMethod(reflect.TypeOf(w), "PowerStatus", func(b *worker) (string, error) {
				return oob.PowerOn, nil
			})

			var ErrPowerOff = errors.New("power off error")
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				So(cmd, ShouldEqual, tool)
				So(len(args), ShouldEqual, 3)
				So(strings.TrimSpace(strings.Join(args, " ")), ShouldEqual, "power off")
				return nil, ErrPowerOff
			})

			defer monkey.UnpatchAll()

			So(NewWorker().PowerOff(), ShouldEqual, ErrPowerOff)
		})
	})
}

func TestPowerReset(t *testing.T) {
	Convey("重启", t, func() {
		Convey("当前已关机并执行开机", func() {
			var w *worker
			monkey.PatchInstanceMethod(reflect.TypeOf(w), "PowerStatus", func(b *worker) (string, error) {
				return oob.PowerOff, nil
			})
			defer monkey.UnpatchAll()

			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				So(cmd, ShouldEqual, tool)
				So(len(args), ShouldEqual, 3)
				So(strings.TrimSpace(strings.Join(args, " ")), ShouldEqual, "power on")
				return nil, nil
			})

			So(NewWorker().PowerReset(), ShouldBeNil)

		})

		Convey("当前已开机并执行重启", func() {
			var w *worker
			monkey.PatchInstanceMethod(reflect.TypeOf(w), "PowerStatus", func(b *worker) (string, error) {
				return oob.PowerOn, nil
			})
			defer monkey.UnpatchAll()

			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				So(cmd, ShouldEqual, tool)
				So(len(args), ShouldEqual, 3)
				So(strings.TrimSpace(strings.Join(args, " ")), ShouldEqual, "power reset")
				return nil, nil
			})

			So(NewWorker().PowerReset(), ShouldBeNil)
		})
	})
}

func TestPXEBoot(t *testing.T) {
	Convey("设备重启并指定其从网络引导", t, func() {
		var count int
		var cmds []string
		var bash *hardware.Bash
		monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
			cmds = append(cmds, fmt.Sprintf("%s %s", cmd, strings.TrimSpace(strings.Join(args, " "))))
			if count == 0 {
				return ioutil.ReadFile("testdata/ipmitool_power_status_on.txt")
			}
			return nil, nil
		})
		defer monkey.UnpatchAll()

		Convey("UEFI", func() {
			count = 0
			cmds = nil
			So(NewWorker().PXEBoot(true), ShouldBeNil)
			So(len(cmds), ShouldEqual, 4)
			for i, expected := range []string{
				"ipmitool power status",
				"ipmitool power off",
				"ipmitool chassis bootdev pxe options=efiboot",
				"ipmitool power on",
			} {
				So(cmds[i], ShouldEqual, expected)
			}
		})

		Convey("Legacy BIOS", func() {
			count = 0
			cmds = nil
			So(NewWorker().PXEBoot(false), ShouldBeNil)
			So(len(cmds), ShouldEqual, 4)
			for i, expected := range []string{
				"ipmitool power status",
				"ipmitool power off",
				"ipmitool chassis bootdev pxe",
				"ipmitool power on",
			} {
				So(cmds[i], ShouldEqual, expected)
			}
		})
	})
}

func TestFRUDevice(t *testing.T) {
	Convey("查询物理机SN、产品名、厂商", t, func() {
		Convey("命令执行无误", func() {
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				return ioutil.ReadFile("./testdata/ipmitool_fru_list_0_dell.txt")
			})
			defer monkey.UnpatchAll()

			fd, err := NewWorker().FRUDevice()
			So(err, ShouldBeNil)
			So(fd, ShouldNotBeNil)
			So(fd.ProductManufacturer, ShouldEqual, "DELL")
			So(fd.ProductName, ShouldEqual, "PowerEdge R620")
			So(fd.ProductSerial, ShouldEqual, "3Q28132")
		})

		Convey("命令执行有误", func() {
			Convey("IP不可达错误", func() {
				var bash *hardware.Bash
				monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
					return []byte("IPMI LAN send command failed\nError: Unable to establish IPMI v2 / RMCP+ session"), oob.NewIPUnreachableError("10.0.106.27", errors.New("hello world"))
				})
				defer monkey.UnpatchAll()

				fd, err := NewWorker().FRUDevice()
				So(oob.IsIPUnreachableError(err), ShouldBeTrue)
				So(fd, ShouldBeNil)
			})

			Convey("用户名、密码不匹配错误", func() {
				var bash *hardware.Bash
				monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
					return []byte("Error: Unable to establish IPMI v2 / RMCP+ session"), oob.NewUsernamePasswordError(errors.New("hello world"))
				})
				defer monkey.UnpatchAll()

				fd, err := NewWorker().FRUDevice()
				So(oob.IsUsernamePasswordError(err), ShouldBeTrue)
				So(fd, ShouldBeNil)
			})

			Convey("Device not present错误", func() {
				var bash *hardware.Bash
				monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
					return []byte("Device not present (Requested sensor, data, or record not found)"), errors.New("Device not present (Requested sensor, data, or record not found)")
				})
				defer monkey.UnpatchAll()

				fd, err := NewWorker().FRUDevice()
				So(oob.IsFRUDeviceNotPresentError(err), ShouldBeTrue)
				So(fd, ShouldBeNil)
			})

			Convey("其它错误", func() {
				execErr := errors.New("exec error")
				var bash *hardware.Bash
				monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
					return nil, execErr
				})
				defer monkey.UnpatchAll()

				fd, err := NewWorker().FRUDevice()
				So(err, ShouldEqual, execErr)
				So(fd, ShouldBeNil)
			})
		})

	})
}

func TestSetDHCP(t *testing.T) {
	Convey("设置带外网络为DHCP", t, func() {
		Convey("远程模式", func() {
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				argsS := strings.Join(args, " ")
				if strings.Contains(argsS, "lan print") {
					return ioutil.ReadFile("./testdata/ipmitool_lan_print_1.txt")
				} else if strings.Contains(argsS, "lan set") {
					return nil, errors.New(tool + " " + strings.Join(args, " "))
				}
				panic("unreachable")
			})
			defer monkey.UnpatchAll()

			err := NewWorker(oob.WithRemote(oob.LANPlusInterface, "10.0.106.27", "root", "calvin")).SetDHCP()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, fmt.Sprintf(`%s -I lanplus -H 10.0.106.27 -U root -P 'calvin' lan set 0 ipsrc dhcp`, tool))
		})

		Convey("本地模式", func() {
			execErr := errors.New("exec error")
			var bash *hardware.Bash
			monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
				return nil, execErr
			})
			defer monkey.UnpatchAll()

			err := NewWorker().SetDHCP()
			So(err, ShouldEqual, oob.ErrChannelNotFound)
		})
	})
}
