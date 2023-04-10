package ipmi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/licairong/cloudboot-provider-framework/oob"
	"github.com/licairong/cloudboot-provider-framework/util"
	"net"
	"strconv"
	"strings"
	"time"

	strutil "github.com/licairong/cloudboot-provider-framework/util/strings"
)

var _ oob.Worker = (*worker)(nil)

const (
	// name 处理器名称
	name = "IPMI"
	// tool 硬件配置工具
	tool = "ipmitool"
)

type worker struct {
	opts     *oob.Options
	log      util.Logger
	executor util.Executor
	shadows  []string // 密码等需要日志脱敏的内容
}

// NewWorker 返回处理器实例
func NewWorker(setters ...func(*oob.Options)) oob.Worker {
	var opts oob.Options
	opts.ChannelID = -1

	for i := range setters {
		setters[i](&opts)
	}

	if opts.Executor == nil {
		opts.Executor = util.NewBash() // 默认的执行器实现
		opts.Executor.SetLog(opts.Log)
	}

	var shadows []string
	if opts.Password != "" {
		shadows = []string{opts.Password}
	}

	return &worker{
		opts:     &opts,
		log:      opts.Log,
		executor: opts.Executor,
		shadows:  shadows,
	}
}

// Name 返回处理器实现的名称
func (w *worker) Name() string {
	return name
}

const (
	lanSendCmdFailed    = "IPMI LAN send command failed"
	unableEstablish     = "Unable to establish"
	fruDeviceNotPresent = "Device not present"
)

// FRUDevice 返回物理机基本信息
func (w *worker) FRUDevice() (fd *oob.FRUDevice, err error) {
	output, err := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "fru", "list", "0")
	if err != nil {
		output, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "fru", "list")
	}
	if err != nil {
		msg := string(output)
		if strings.Contains(msg, lanSendCmdFailed) && strings.Contains(msg, unableEstablish) {
			var oobip string
			if w != nil && w.opts != nil {
				oobip = w.opts.Hostname
			}
			return nil, oob.NewIPUnreachableError(oobip, err)
		}
		if strings.Contains(msg, unableEstablish) {
			return nil, oob.NewUsernamePasswordError(err)
		}
		if strings.Contains(msg, fruDeviceNotPresent) {
			return nil, oob.NewFRUDeviceNotPresentError("0", err)
		}
		return nil, err
	}
	return w.parseFRUDevice(output)
}

// ValidateSN 校验预期的SN与实际的SN是否匹配
func (w *worker) ValidateSN(sn string) (err error) {
	fd, err := w.FRUDevice()
	if err != nil {
		return err
	}
	if fd.ProductSerial != sn {
		return oob.ErrOOBIPAndSNUnmatched
	}
	return nil
}

// 返回电源状态
func (w *worker) PowerStatus() (status string, err error) {
	out, err := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "power", "status")
	if err != nil {
		return "", err
	}
	if strings.Contains(string(out), "Chassis Power is on") {
		return oob.PowerOn, nil
	}
	if strings.Contains(string(out), "Chassis Power is off") {
		return oob.PowerOff, nil
	}
	return oob.PowerOff, nil
}

// 设备上电开机
func (w *worker) PowerOn() (err error) {
	if status, _ := w.PowerStatus(); status == oob.PowerOn {
		return nil
	}
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "power", "on")
	w.Sleep()
	return err
}

// 设备下电关机
func (w *worker) PowerOff() (err error) {
	if status, _ := w.PowerStatus(); status == oob.PowerOff {
		return nil
	}
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "power", "off")
	w.Sleep()
	return err
}

// 设备重启
func (w *worker) PowerReset() (err error) {
	if status, _ := w.PowerStatus(); status == oob.PowerOff {
		w.Sleep()
		return w.PowerOn() // 关机状态下无法直接重启
	}
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "power", "reset")
	return err
}

// 设备重启并指定其从网络引导
func (w *worker) PXEBoot(uefi bool, manufacturer string) (err error) {
	_ = w.PowerOff()
	w.Sleep()

	args := []string{w.remoteArgs(), "chassis", "bootdev", "pxe"}
	if uefi {
		//海通Sugon R6230HA型号服务器不能成功从pxe启动，故注释掉 最初是因为翼支付曙光I620-G20不能成功启动添加
		// if manufacturer == hardware.Sugon || manufacturer == hardware.Suma {
		// 	args = append(args, "options=efiboot,persistent")
		// } else {
		args = append(args, "options=efiboot")
		//}

	}
	if _, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, args...); err != nil {
		return err
	}
	w.Sleep()
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "power", "on")
	return err
}

// Raw 发送原始的ipmi请求并返回响应内容
func (w *worker) Raw(args string) (response []byte, err error) {
	if strings.ContainsAny(args, ",;|><&$!") {
		return nil, errors.New("invalid commandline args")
	}
	return w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, "raw", args)
}

// SetDHCP 设置IP来源是DHCP
func (w *worker) SetDHCP() error {
	channel, err := w.getBuffedChannel()
	if err != nil {
		return err
	}
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "lan", "set", strconv.Itoa(channel), "ipsrc", "dhcp")
	return err
}

// SetStaticIP 设置IP来源是静态IP
func (w *worker) SetStaticIP(ip, netmask, gateway string) error {
	channel, err := w.getBuffedChannel()
	if err != nil {
		return err
	}
	ch := strconv.Itoa(channel)
	if _, err_new := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "lan", "set", ch, "ipsrc", "static"); err_new != nil {
		if err != nil {
			err = errors.New(err.Error() + ";" + err_new.Error())
		} else {
			err = err_new
		}
	}
	if _, err_new := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "lan", "set", ch, "ipaddr", ip); err_new != nil {
		if err != nil {
			err = errors.New(err.Error() + ";" + err_new.Error())
		} else {
			err = err_new
		}
	}
	if _, err_new := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "lan", "set", ch, "netmask", netmask); err_new != nil {
		if err != nil {
			err = errors.New(err.Error() + ";" + err_new.Error())
		} else {
			err = err_new
		}
	}
	if _, err_new := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "lan", "set", ch, "defgw", "ipaddr", gateway); err_new != nil {
		if err != nil {
			err = errors.New(err.Error() + ";" + err_new.Error())
		} else {
			err = err_new
		}
	}
	return err
}

// ChangeUserPassword 修改目标带外用户密码
func (w *worker) ChangeUserPassword(username, password string) (err error) {
	user, err := w.findUserByName(username)
	if err != nil {
		return err
	}
	// ipmitool user set password $userid "$_pw"
	shadows := make([]string, 0, len(w.shadows)+1)
	shadows = append(shadows, w.shadows...)
	shadows = append(shadows, password)
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: shadows}, tool, w.remoteArgs(), "user", "set", "password", strconv.Itoa(user.ID), fmt.Sprintf("%q", password))
	return err
}

// GenerateUser 生成用户带外帐号
func (w *worker) GenerateUser(sett *oob.UserSettingItem) error {
	var channel, userID int
	user, err := w.findUserByName(sett.Username)
	if err != nil && !oob.IsUserNotFoundError(err) {
		return err
	}
	if user == nil { // 目标用户不存在
		channel, err = w.getBuffedChannel()
		if err != nil {
			return err
		}

		userID, err = w.newUserID()
		if err != nil {
			return err
		}

		// ipmitool user set name $userid "$_user"
		if _, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "user", "set", "name", strconv.Itoa(userID), fmt.Sprintf("%q", sett.Username)); err != nil {
			return err
		}
		time.Sleep(500 * time.Millisecond)

	} else { // 目标用户已存在
		channel = user.Channel
		userID = user.ID
	}

	// ipmitool user set password $userid "$_pw"
	w.shadows = append(w.shadows, sett.Password)
	if _, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "user", "set", "password", strconv.Itoa(userID), fmt.Sprintf("%q", sett.Password)); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	// ipmitool user enable $userid
	if _, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "user", "enable", strconv.Itoa(userID)); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	// ipmitool user priv $userid $privilege_level $channel
	var output []byte
	if output, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "user", "priv", strconv.Itoa(userID), strconv.Itoa(sett.PrivilegeLevel), strconv.Itoa(channel)); err != nil {
		err = errors.New(string(output))
		return err
	}
	time.Sleep(500 * time.Millisecond)

	// ipmitool channel setaccess $channel $userid callin=on ipmi=on link=on privilege=$privilege_level
	if _, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "channel", "setaccess", strconv.Itoa(channel), strconv.Itoa(userID), "callin=on", "ipmi=on", "link=on", fmt.Sprintf("privilege=%d", sett.PrivilegeLevel)); err != nil {
		// TODO 暂时性处理办法：去掉'link=on'后重试
		if _, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "channel", "setaccess", strconv.Itoa(channel), strconv.Itoa(userID), "callin=on", "ipmi=on", fmt.Sprintf("privilege=%d", sett.PrivilegeLevel)); err != nil {
			return err
		}
		return err
	}

	// 尝试开启iDRAC访问权限
	_ = w.enableUserAccessIDRAC(userID)

	switch sett.Status {
	case oob.EnabledUser:
		return w.enableUser(userID)
	case oob.DisabledUser:
		return w.disableUser(userID)
	}
	return nil
}

// EnableUser 启用带外用户帐号
func (w *worker) EnableUser(username string) (err error) {
	user, err := w.findUserByName(username)
	if err != nil {
		return err
	}
	return w.enableUser(user.ID)
}

// DisableUser 禁用带外用户帐号
func (w *worker) DisableUser(username string) error {
	user, err := w.findUserByName(username)
	if err != nil {
		return err
	}
	return w.disableUser(user.ID)
}

// Users 返回OOB用户列表
func (w *worker) Users() ([]*oob.User, error) {
	channel, err := w.getBuffedChannel()
	if err != nil {
		return nil, err
	}

	output, err := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "user", "list", strconv.Itoa(channel))
	if err != nil {
		return nil, err
	}

	users, err := parseUsers(output)
	if err != nil {
		return nil, err
	}

	for i := range users {
		users[i].Channel = channel
		users[i].Access, _ = w.userAccess(channel, users[i].ID)
		if users[i].Access != nil && users[i].Access.UserName != "" {
			users[i].Name = users[i].Access.UserName // 解析ipmitool user list $channel的输出可能无法得到正确的用户名
		}
	}
	return users, nil
}

// BMCColdReset (冷)重启BMC
func (w *worker) BMCColdReset() error {
	_, _ = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "mc", "reset", "cold")
	return nil // 假设每次执行'ipmitool mc reset cold'都能达到预期效果，丢弃error。
}

// channel 返回channel。
func (w *worker) Channel() (int, error) {
	// pairs := make(map[int]string, 1)
	channel := -1
	// 试错法查找channel
	for i := 0; i <= 10; i++ {
		out, _ := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "lan", "print", strconv.Itoa(i))
		if len(out) <= 0 || strings.Contains(string(out), "Invalid channel") { // 通过exit code判断并不能保证一定准确
			continue
		}
		if n, _ := w.parseNetwork(out); n != nil && n.IP != "" && n.IP != "0.0.0.0" {
			return i, nil
		}
		if channel < 0 {
			channel = i
		}
	}
	if channel < 0 {
		return 0, oob.ErrChannelNotFound
	}
	if w.opts == nil {
		w.opts = new(oob.Options)
	}
	w.opts.ChannelID = channel
	return channel, nil
}

// BMC 返回OOB的BMC信息
func (w *worker) BMC() (*oob.BMC, error) {
	output, err := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "mc", "info")
	if err != nil {
		return nil, err
	}

	var bmc oob.BMC
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Firmware Revision") {
			bmc.FirmwareReversion = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "IPMI Version") {
			bmc.IPMIVersion = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "Manufacturer ID") {
			bmc.ManufacturerID = strutil.ExtractValue(line, strutil.ColonSep)
		} else if strings.HasPrefix(line, "Manufacturer Name") {
			bmc.ManufacturerName = strutil.ExtractValue(line, strutil.ColonSep)
		}
	}
	return &bmc, nil
}

// Network 返回OOB网络信息
func (w *worker) Network() (*oob.Network, error) {
	channel, err := w.getBuffedChannel()
	if err != nil {
		return nil, err
	}

	output, _ := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "lan", "print", strconv.Itoa(channel)) // 舍弃error。因为部分机型下执行该命令，即使命令输出正确，exit code也会是一个非0值。
	return w.parseNetwork(output)
}

// PostCheck OOB配置实施后置检查
func (w *worker) PostCheck(sett *oob.Setting) (items []*util.CheckingItem) {
	if sett == nil {
		return nil
	}
	if sett.Network != nil {
		items = append(items, w.checkNetwork(sett.Network)...)
	}
	if sett.User != nil {
		items = append(items, w.checkUser(sett.User)...)
	}
	return items
}

// BMC重启
func (w *worker) BmcReset() (err error) {
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "power", "reset")
	return err
}

// SelClear 日志清空
func (w *worker) SelClear() error {
	_, err := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "sel", "clear")
	return err
}

// SensorList 返回物理机传感器信息。
func (w *worker) SensorList() (sensorlist []*oob.SensorDevice, err error) {
	output, err := w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "sensor", "list")
	if err != nil {
		msg := string(output)
		if strings.Contains(msg, lanSendCmdFailed) && strings.Contains(msg, unableEstablish) {
			var oobip string
			if w != nil && w.opts != nil {
				oobip = w.opts.Hostname
			}
			return nil, oob.NewIPUnreachableError(oobip, err)
		}
		if strings.Contains(msg, unableEstablish) {
			return nil, oob.NewUsernamePasswordError(err)
		}
		return nil, err
	}
	return w.parseSensorlist(output)
}

// SensorList 返回物理机传感器信息。
func (w *worker) SetSnmpTrap(snmpset *oob.SnmpSet) (err error) {
	var version, alarmseverity, authProtocol, privProtocol string
	switch snmpset.SnmpTrapVersion {
	case "1":
		version = "0x01"
	case "2c":
		version = "0x02"
	case "3":
		version = "0x03"
	}
	if version == "" {
		return errors.New("版本不合法 合法值为1/2c/3")
	}
	//设置告警版本
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x00", version)
	if err != nil {
		return err
	}
	switch snmpset.SnmpTrapAlarmseverity {
	case "critical":
		alarmseverity = "0x02"
	case "warning":
		alarmseverity = "0x01"
	case "all":
		alarmseverity = "0x00"
	}
	if alarmseverity == "" {
		return errors.New("告警Event等级不合法 合法值为critical/warning/all")
	}
	//设置告警Event等级
	_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x0d", alarmseverity)
	if err != nil {
		return err
	}
	//设置团体名
	if snmpset.SnmpTrapVersion == "1" || snmpset.SnmpTrapVersion == "2c" {
		if snmpset.CommunityName == "" {
			return errors.New("团体名不合法 不能为空")
		}
		//设置团体名
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x01", fmt.Sprintf(`string2HexData "设置团体名" "%s" 161`, snmpset.CommunityName))
		if err != nil {
			return err
		}
		//清空用户名
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3C 0x19 0x02", "0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00 0x00")
	} else {
		if snmpset.SnmpV3User == "" || snmpset.SnmpV3AuthPassword == "" || snmpset.SnmpV3PrivPassword == "" {
			return errors.New("用户名 必需参数用户名或认证密码或加密密码为空")
		}
		switch snmpset.SnmpV3AuthProtocol {
		case "SHA":
			authProtocol = "0x01"
		case "MD5":
			authProtocol = "0x02"
		}
		if authProtocol == "" {
			return errors.New("告警认证协议不合法 可选SHA或MD5")
		}
		switch snmpset.SnmpV3PrivProtocol {
		case "DES":
			privProtocol = "0x01"
		case "AES":
			privProtocol = "0x02"
		}
		if privProtocol == "" {
			return errors.New("告警加密协议不合法 可选DES或AES")
		}
		//设置告警用户名
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3C 0x19 0x02", fmt.Sprintf(`string2HexData "设置告警用户名%s" "%s" 32`, snmpset.SnmpV3User, snmpset.SnmpV3User))
		if err != nil {
			return err
		}
		//设置告警认证协议
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x03", authProtocol)
		if err != nil {
			return err
		}
		//设置告警认证密码
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x05", fmt.Sprintf(`string2HexData "设置告警认证密码" "%s" 161`, snmpset.SnmpV3AuthPassword))
		if err != nil {
			return err
		}
		//设置告警加密协议
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x04", privProtocol)
		if err != nil {
			return err
		}
		//设置告警加密密码
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x06", fmt.Sprintf(`string2HexData "转换加密密码" "%s" 161`, snmpset.SnmpV3PrivPassword))
		if err != nil {
			return err
		}
		if snmpset.SnmpTrapEngineId > 0 {
			//设置告警引擎号EngineID
			_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x07", fmt.Sprintf(`string2HexData "设置告警引擎ID%d" "%d" 48`, snmpset.SnmpTrapEngineId, snmpset.SnmpTrapEngineId))
			if err != nil {
				return err
			}
		}
	}
	if snmpset.SnmpTrapSystemName != "" {
		//设置告警系统名
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x0b", fmt.Sprintf(`string2HexData "设置告警系统名%s" "%s" 32`, snmpset.SnmpTrapSystemName, snmpset.SnmpTrapSystemName))
		if err != nil {
			return err
		}
	}
	if snmpset.SnmpTrapSystemId > 0 {
		//设置告警系统ID
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x0c", fmt.Sprintf(`string2HexData "设置告警系统ID%d" "%d" 16`, snmpset.SnmpTrapSystemId, snmpset.SnmpTrapSystemId))
		if err != nil {
			return err
		}
	}
	if snmpset.SnmpTrapLocation != "" {
		//设置告警主机位置
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x09", fmt.Sprintf(`string2HexData "设置告警主机位置%s" "%s" 48`, snmpset.SnmpTrapLocation, snmpset.SnmpTrapLocation))
		if err != nil {
			return err
		}
	}
	if snmpset.SnmpTrapContact != "" {
		//设置告警联系人
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x08", fmt.Sprintf(`string2HexData "设置告警联系人%s" "%s" 16`, snmpset.SnmpTrapContact, snmpset.SnmpTrapContact))
		if err != nil {
			return err
		}
	}
	if snmpset.SnmpTrapHostOs != "" {
		//设置告警系统主机名
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x3c 0x19 0x0a", fmt.Sprintf(`string2HexData "设置告警系统主机名%s" "%s" 16`, snmpset.SnmpTrapHostOs, snmpset.SnmpTrapHostOs))
		if err != nil {
			return err
		}
	}
	if snmpset.SnmpTrapPortNo > 0 {
		raw := "raw 0x3c 0x19 0x0e"
		if strings.Contains(snmpset.Devicemodel, "NF8260M5") || strings.Contains(snmpset.Devicemodel, "NF8460M5") || strings.Contains(snmpset.Devicemodel, "NF8480M5") {
			raw = "raw 0x3a 0x08"
		}
		//设置告警端口号
		_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), raw, fmt.Sprintf("%d", snmpset.SnmpTrapPortNo))
		if err != nil {
			return err
		}
	}
	for _, v := range snmpset.SnmpTrapServer {
		if v.SnmpTrapPolicy == "disable" {
			//关闭告警策略
			_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x04 0x12 0x09", fmt.Sprintf("0x0%s", v.TrapID), fmt.Sprintf("0x%s0", v.TrapID), fmt.Sprintf("0x1%s", v.TrapID), "0x00")
			if err != nil {
				return err
			}
		}
		if v.SnmpTrapPolicy == "enable" {
			//开启告警策略
			_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x04 0x12 0x09", fmt.Sprintf("0x0%s", v.TrapID), fmt.Sprintf("0x%s8", v.TrapID), fmt.Sprintf("0x%s%s", v.SnmpTrapChannel, v.TrapID), "0x00")
			if err != nil {
				return err
			}
		}
		trapType := ""
		switch v.SnmpTrapType {
		case "email":
			trapType = ""
		case "snmp":
			trapType = "00"
		}
		if trapType != "" {
			//设置告警类型
			_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x0c 0x01", fmt.Sprintf("0x0%s 0x12 ", v.SnmpTrapChannel), fmt.Sprintf("0x0%s", v.TrapID), fmt.Sprintf(" 0x%s 0x03 0x03", trapType))
			if err != nil {
				return err
			}
		}
		if v.SnmpTrapDestination != "" {
			b_ip4 := net.ParseIP(v.SnmpTrapDestination).To4()
			if b_ip4 != nil {
				_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x0c 0x01", fmt.Sprintf("0x0%s 0xC1 ", v.SnmpTrapChannel), fmt.Sprintf(" 0x0%s 0x01 0x00", v.TrapID), fmt.Sprintf("%s 0x00 0x00 0x00 0x00 0x00 0x00", v.SnmpTrapDestination))
				if err != nil {
					return err
				}
			}
			b_ip6 := net.ParseIP(v.SnmpTrapDestination).To16()
			if b_ip6 != nil {
				_, err = w.executor.Exec(&util.ExecutionOptions{Shadows: w.shadows}, tool, w.remoteArgs(), "raw 0x0c 0x01", fmt.Sprintf("0x0%s 0x13 ", v.SnmpTrapChannel), fmt.Sprintf(" 0x0%s 0x00 0x00", v.TrapID), fmt.Sprintf("%s 0x00 0x00 0x00 0x00 0x00 0x00", v.SnmpTrapDestination))
				if err != nil {
					return err
				}
			}
		}
	}
	return err
}
