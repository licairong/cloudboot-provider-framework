package ipmi

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/licairong/cloudboot-provider-framework/oob"
	"github.com/licairong/cloudboot-provider-framework/util"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	strutil "github.com/licairong/cloudboot-provider-framework/util/strings"
)

// Sleep 休眠2s
func (w *worker) Sleep() {
	time.Sleep(3 * time.Second)
}

// remoteArgs 返回发起远程命令所使用的命令行选项字符串
func (w *worker) remoteArgs() string {
	if w == nil || w.opts == nil || w.opts.Interface == "" || w.opts.Hostname == "" || w.opts.Username == "" {
		return ""
	}
	return fmt.Sprintf("-I %s -H %s -U %s -P '%s'",
		w.opts.Interface,
		w.opts.Hostname,
		w.opts.Username,
		w.opts.Password,
	)
}

func (w *worker) getBuffedChannel() (channel int, err error) {
	if w.opts != nil && w.opts.ChannelID >= 0 {
		return w.opts.ChannelID, nil
	}
	return w.Channel()
}

// EnableUser 启用带外用户帐号
func (w *worker) enableUser(userID int) (err error) {
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "user", "enable", strconv.Itoa(userID))
	return err
}

// DisableUser 禁用带外用户帐号
func (w *worker) disableUser(userID int) (err error) {
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "user", "disable", strconv.Itoa(userID))
	return err
}

// findUserByName 根据带外用户名查找带外用户。
// 若指定名称的用户不存在，则返回oob.UserNotFoundError实例。
func (w *worker) findUserByName(name string) (*oob.User, error) {
	users, err := w.Users()
	if err != nil {
		return nil, err
	}
	for i := range users {
		if users[i] == nil {
			continue
		}
		if users[i].Name == name {
			return users[i], nil
		}
	}
	return nil, oob.NewUserNotFoundError(name)
}

// findUserIndexByName 返回指定用户名的用户在切片中的索引号
func (w *worker) findUserIndexByName(users []*oob.User, name string) (index int) {
	for i := range users {
		if users[i] == nil {
			continue
		}
		if users[i].Name == name {
			return i
		}
	}
	return -1
}

const (
	defaultUserID = 2
)

// newUserID 返回新建用户的ID
func (w *worker) newUserID() (id int, err error) {
	users, err := w.Users()
	if err != nil {
		return 0, err
	}
	if len(users) <= 0 {
		return defaultUserID, nil
	}

	for i := range users {
		name := strings.TrimSpace(users[i].Name)
		if name == "" || strings.Contains(name, "(Empty") { // 系统预置用户
			return users[i].ID, nil
		}
	}
	return users[len(users)-1].ID + 1, nil
}

// userAccess 返回通道下指定用户的Access信息
func (w *worker) userAccess(channel, userID int) (*oob.UserAccess, error) {
	output, err := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "channel", "getaccess", strconv.Itoa(channel), strconv.Itoa(userID))
	if err != nil {
		return nil, err
	}
	var access oob.UserAccess
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "User Name") {
			access.UserName = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "Access Available") {
			access.AccessAvailable = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "Link Authentication") {
			access.LinkAuthentication = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "IPMI Messaging") {
			access.IPMIMessaging = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "Privilege Level") {
			access.PrivilegeLevel = oob.IntUserLevel(strutil.ExtractValue(line, strutil.ColonSep))
		}
	}
	return &access, nil
}

func (w *worker) parseFRUDevice(output []byte) (*oob.FRUDevice, error) {
	var fd oob.FRUDevice
	var chassis string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if fd.ProductManufacturer == "" && strings.HasPrefix(line, "Product Manufacturer") {
			fd.ProductManufacturer = strutil.ExtractValue(line, strutil.ColonSep)

		} else if fd.ProductName == "" && strings.HasPrefix(line, "Product Name") {
			fd.ProductName = strutil.ExtractValue(line, strutil.ColonSep)

		} else if fd.ProductSerial == "" && strings.HasPrefix(line, "Product Serial") {
			fd.ProductSerial = strutil.ExtractValue(line, strutil.ColonSep)

		} else if strings.HasPrefix(line, "Chassis Serial") {
			chassis = strutil.ExtractValue(line, strutil.ColonSep)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if fd.ProductSerial == "" && chassis != "" {
		fd.ProductSerial = chassis
	}
	return &fd, nil
}

func (w *worker) parseNetwork(output []byte) (*oob.Network, error) {
	var network oob.Network
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "IP Address Source") {
			network.IPSrc = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "IP Address") {
			network.IP = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "Subnet Mask") {
			network.Netmask = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "MAC Address") {
			network.MAC = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "Default Gateway IP") {
			network.Gateway = strutil.ExtractValue(line, strutil.ColonSep)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &network, nil
}

var (
	numReg = regexp.MustCompile("^\\d+$")
)

// parseHardDisks 返回stdout内容中的带外用户信息
func parseUsers(output []byte) (items []*oob.User, err error) {
	var started bool
	rd := bufio.NewReader(bytes.NewBuffer(output))
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ID") && strings.Contains(line, "Name") {
			started = true
			continue
		}
		if !started {
			continue
		}

		arr := strings.Fields(line)
		if len(arr) < 6 || !numReg.MatchString(arr[0]) || arr[1] == "true" || arr[1] == "false" || arr[1] == "(Empty" {
			continue
		}

		id, _ := strconv.Atoi(arr[0])
		items = append(items, &oob.User{
			ID:   id,
			Name: arr[1],
		})
	}
	return items, nil
}

// isDELL 返回当前设备的厂商是否是DELL的布尔值
func (w *worker) isDELL() bool {
	fd, err := w.FRUDevice()
	if err != nil || fd == nil {
		return true
	}
	return util.ManufacturerName(fd.ProductManufacturer) == util.Dell
}

// racadmTool DELL iDRAC配置工具
const racadmTool = "/opt/dell/srvadmin/sbin/racadm"

func (w *worker) enableUserAccessIDRAC(userID int) (err error) {
	if !w.isDELL() {
		return nil
	}
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, racadmTool, "config", "-g", "cfgUserAdmin", "-i", strconv.Itoa(userID), "-o", "cfgUserAdminPrivilege", "0x000001ff")
	return err
}
func (w *worker) parseSensorlist(output []byte) ([]*oob.SensorDevice, error) {
	items := make([]*oob.SensorDevice, 0)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Split(line, "|")
		if len(fields) != 10 {
			continue
		}
		for k, v := range fields {
			if strings.TrimSpace(v) == "na" || strings.TrimSpace(v) == "0x0" {
				fields[k] = ""
			}
		}
		items = append(items, &oob.SensorDevice{
			Name:     strings.TrimSpace(fields[0]),
			Value:    strings.TrimRight(strings.TrimRight(strings.TrimSpace(fields[1]), "0"), "."),
			Units:    strings.TrimSpace(fields[2]),
			State:    strings.TrimSpace(fields[3]),
			Lonorec:  strings.TrimRight(strings.TrimRight(strings.TrimSpace(fields[4]), "0"), "."),
			Locrit:   strings.TrimRight(strings.TrimRight(strings.TrimSpace(fields[5]), "0"), "."),
			Lonocrit: strings.TrimRight(strings.TrimRight(strings.TrimSpace(fields[6]), "0"), "."),
			Upcrit:   strings.TrimRight(strings.TrimRight(strings.TrimSpace(fields[7]), "0"), "."),
			Upnocrit: strings.TrimRight(strings.TrimRight(strings.TrimSpace(fields[8]), "0"), "."),
			Upnorec:  strings.TrimRight(strings.TrimRight(strings.TrimSpace(fields[9]), "0"), "."),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
