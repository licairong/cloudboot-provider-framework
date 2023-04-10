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

func readFile(cmdArgs string) ([]byte, error) {
	switch strings.Join(strings.Fields(cmdArgs), " ") {
	case fmt.Sprintf("%s lan print 0", tool):
		return nil, errors.New("exec error")
	case fmt.Sprintf("%s lan print 1", tool):
		return ioutil.ReadFile("./testdata/ipmitool_lan_print_1.txt")
	case fmt.Sprintf("%s user list 1", tool):
		return ioutil.ReadFile("./testdata/ipmitool_user_list_1.txt")
	case fmt.Sprintf("%s channel getaccess 1 2", tool):
		return ioutil.ReadFile("./testdata/ipmitool_channel_getaccess_1_2.txt")
	case fmt.Sprintf("%s channel getaccess 1 6", tool):
		return ioutil.ReadFile("./testdata/ipmitool_channel_getaccess_1_6.txt")
	default:
		fmt.Println("panic", "==>", cmdArgs)
		panic("unreachable")
	}
}

func Test_findUserByName(t *testing.T) {
	Convey("根据带外用户名查找用户", t, func() {
		var bash *hardware.Bash

		monkey.PatchInstanceMethod(reflect.TypeOf(bash), "Exec", func(b *hardware.Bash, opts *hardware.ExecutionOptions, cmd string, args ...string) ([]byte, error) {
			cmdArgs := make([]string, 0, len(args)+1)
			cmdArgs = append(cmdArgs, cmd)
			cmdArgs = append(cmdArgs, args...)
			return readFile(strings.Join(cmdArgs, " "))
		})
		defer monkey.UnpatchAll()

		w := new(worker)
		w.executor = bash

		user, err := w.findUserByName("root")
		So(err, ShouldBeNil)
		So(user, ShouldNotBeNil)
		So(user.ID, ShouldEqual, 2)
		So(user.Name, ShouldEqual, "root")

		user, err = w.findUserByName("voidint")
		So(err, ShouldBeNil)
		So(user, ShouldNotBeNil)
		So(user.ID, ShouldEqual, 6)
		So(user.Name, ShouldEqual, "voidint")

		user, err = w.findUserByName("not_exist")
		So(err, ShouldNotBeNil)
		So(oob.IsUserNotFoundError(err), ShouldBeTrue)
		So(user, ShouldBeNil)
	})
}

func Test_findUserIndexByName(t *testing.T) {
	Convey("根据用户名返回元素在切片中的下标", t, func() {
		users := []*oob.User{
			{Name: "tangtianyun"},
			{Name: "voidint"},
			{Name: "tty"},
		}
		So(new(worker).findUserIndexByName(users, "voidint"), ShouldEqual, 1)
		So(new(worker).findUserIndexByName(users, "hello"), ShouldEqual, -1)
	})
}

func Test_newUserID(t *testing.T) {
	Convey("返回新带外用户ID", t, func() {
		var w *worker
		monkey.PatchInstanceMethod(reflect.TypeOf(w), "Users", func(*worker) ([]*oob.User, error) {
			return []*oob.User{
				{Channel: 1, ID: 1, Name: "Administrator"},
				{Channel: 1, ID: 2, Name: "golang"},
				{Channel: 1, ID: 3, Name: "(Empty User)"},
				{Channel: 1, ID: 4, Name: "(Empty User)"},
				{Channel: 1, ID: 5, Name: "(Empty User)"},
				{Channel: 1, ID: 6, Name: "(Empty User)"},
			}, nil
		})
		defer monkey.UnpatchAll()

		id, err := w.newUserID()
		So(err, ShouldBeNil)
		So(id, ShouldEqual, 3)
	})
}

func Test_parseUsers(t *testing.T) {
	Convey("解析命令行输出并提取其中包含的带外用户信息", t, func() {
		output, err := ioutil.ReadFile("./testdata/ipmitool_user_list_1.txt")
		So(err, ShouldBeNil)

		items, err := parseUsers(output)
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 2)

		So(items[0].ID, ShouldEqual, 2)
		So(items[0].Name, ShouldEqual, "root")

		So(items[1].ID, ShouldEqual, 6)
		So(items[1].Name, ShouldEqual, "voidint")

		output, err = ioutil.ReadFile("./testdata/ipmitool_user_list_with_empty_users.txt")
		So(err, ShouldBeNil)

		items, err = parseUsers(output)
		So(err, ShouldBeNil)
		So(len(items), ShouldEqual, 2)

		So(items[0].ID, ShouldEqual, 1)
		So(items[0].Name, ShouldEqual, "Administrator")

		So(items[1].ID, ShouldEqual, 2)
		So(items[1].Name, ShouldEqual, "golang")
	})
}

func Test_remoteArgs(t *testing.T) {
	Convey("ipmitool远程模式命令行选项", t, func() {
		Convey("远程模式命令行选项", func() {
			oobw := NewWorker(oob.WithRemote(oob.LANPlusInterface, "10.0.106.27", "root", "calvin"))
			So(oobw, ShouldNotBeNil)
			w, ok := oobw.(*worker)
			So(ok, ShouldBeTrue)
			So(w, ShouldNotBeNil)
			So(w.remoteArgs(), ShouldEqual, `-I lanplus -H 10.0.106.27 -U root -P 'calvin'`)
		})

		Convey("本地模式命令行选项", func() {
			oobw := NewWorker()
			So(oobw, ShouldNotBeNil)
			w, ok := oobw.(*worker)
			So(ok, ShouldBeTrue)
			So(w, ShouldNotBeNil)
			So(w.remoteArgs(), ShouldBeBlank)
		})
	})
}

func Test_parseFRUDevice(t *testing.T) {
	Convey("解析FRU Device", t, func() {
		Convey("DELL", func() {
			output, err := ioutil.ReadFile("./testdata/ipmitool_fru_list_0_dell.txt")
			So(err, ShouldBeNil)
			So(output, ShouldNotBeNil)
			fd, err := new(worker).parseFRUDevice(output)
			So(err, ShouldBeNil)
			So(output, ShouldNotBeNil)
			So(fd.ProductManufacturer, ShouldEqual, "DELL")
			So(fd.ProductName, ShouldEqual, "PowerEdge R620")
			So(fd.ProductSerial, ShouldEqual, "3Q28132")
		})

		Convey("HP(ipmitool fru list 0)", func() {
			output, err := ioutil.ReadFile("./testdata/ipmitool_fru_list_0_hp.txt")
			So(err, ShouldBeNil)
			So(output, ShouldNotBeNil)
			fd, err := new(worker).parseFRUDevice(output)
			So(err, ShouldBeNil)
			So(output, ShouldNotBeNil)
			So(fd.ProductManufacturer, ShouldEqual, "HP")
			So(fd.ProductName, ShouldEqual, "ProLiant DL360e Gen8")
			So(fd.ProductSerial, ShouldEqual, "CN722903GJ")
		})

		Convey("HP(ipmitool fru list)", func() {
			output, err := ioutil.ReadFile("./testdata/ipmitool_fru_list_hp.txt")
			So(err, ShouldBeNil)
			So(output, ShouldNotBeNil)
			fd, err := new(worker).parseFRUDevice(output)
			So(err, ShouldBeNil)
			So(output, ShouldNotBeNil)
			So(fd.ProductManufacturer, ShouldEqual, "HP")
			So(fd.ProductName, ShouldEqual, "ProLiant DL360e Gen8")
			So(fd.ProductSerial, ShouldEqual, "CN722903GJ")
		})
	})
}
