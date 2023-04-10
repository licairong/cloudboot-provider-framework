package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/licairong/cloudboot-provider-framework/collector"
	"github.com/licairong/cloudboot-provider-framework/shared"
	"github.com/licairong/cloudboot-provider-framework/util"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	//"github.com/astaxie/beego/httplib"
)

var (
	base         *collector.Base
	plugin_name  string
	protocol     plugin.ClientProtocol
	meta         *Meta
	dev          collector.Device
	executor = util.NewBash()
)

type Meta struct {
	Manufacturer string    `json:"manufacturer"`
	Model        string    `json:"model"`
	Raid         *Raid     `json:"raid"`
	Bios         *Hardware `json:"bios"`
	Oob          *Hardware `json:"oob"`
}

type Raid struct {
	Manufacturer     string `json:"manufacturer"`
	Model            string `json:"model"`
	Raid             *Raid  `json:"raid"`
	Tool             string `json:"tool"`
	ToolVersion      string `json:"tool_version"`
	ToolDownloadAddr string `json:"tool_download_addr"`
}

type Hardware struct {
	Version          string `json:"version"`
	Tool             string `json:"tool"`
	ToolVersion      string `json:"tool_version"`
	ToolDownloadAddr string `json:"tool_download_addr"`
}

func handleErr(err error, exit bool) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "Error ==> %s\n", err.Error())
	if exit {
		os.Exit(1)
	}
}

func SelfCheck() {
	base, _ = collector.BASE()
	data, err := json.MarshalIndent(base, "", "    ")
	handleErr(err, false)
	fmt.Println("===============Base===============", "\n", string(data))
	if base.Manufacturer == "" || base.Model == "" || base.SN == "" {
		fmt.Println("Error: device info not collect, exit.")
		//os.Exit(1)
	}
	plugin_name = strings.Replace(fmt.Sprintf(base.Manufacturer, "-", base.Model), " ", "-", 1)
}

func LoadProvider() {
	// 下载 provider
	//tmp_file := "/tmp/plugin_file"
	//url := fmt.Sprintf("%s/plugins/%s_%s.zip", agent.Cmdline.ServerAddr, plugin_name, base.Arch)
	//if err := httplib.Get(url).ToFile(tmp_file); err != nil {
	//	fmt.Println("download provider error: %s", err)
	//	os.Exit(1)
	//}
	//
	//// 解压 provider 到 /usr/bin
	//if err := util.UnZip("/usr/bin", tmp_file); err != nil {
	//	fmt.Println("unzip provider error: %s", err)
	//	os.Exit(1)
	//}

	// 加载 provider
	pluginClientConfig := &plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Cmd:              exec.Command("./dell-poweredge-r730"),
		Plugins:          map[string]plugin.Plugin{
			"raid":       &shared.GRPCRaidPlugin{},
			"oob":        &shared.GRPCOobPlugin{},
		},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	}

	client := plugin.NewClient(pluginClientConfig)
	pluginClientConfig.Reattach = client.ReattachConfig()
	protocol, _ = client.Client()
}

func InstallTools() {
	// 加载元数据
	loadMeta()

	// 安装硬件配置工具
	_, err := executor.Exec(nil, "yum",  "-y", "install",  meta.Raid.Tool)
	if err != nil {
		fmt.Println("install raid tool failed,", err)
	}

	_, err = executor.Exec(nil, "yum",  "-y", "install",  meta.Bios.Tool)
	if err != nil {
		fmt.Println("install bios tool failed,", err)
	}

	_, err = executor.Exec(nil, "yum",  "-y", "install",  meta.Oob.Tool)
	if err != nil {
		fmt.Println("install oob tool failed,", err)
	}
}

func loadMeta() {
	jsonFile, err := os.Open("meta.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	_ = json.Unmarshal([]byte(byteValue), &meta)
}

// CollectDeviceInfo 采集设备信息并丢弃采集过程中的错误。
func CollectDeviceInfo() {
	defer func() {
		if re := recover(); re != any(nil) {
			fmt.Println("Panic: %v: %s", re, debug.Stack())
		}
	}()

	dev.IsVM = base.IsVM
	dev.SN = base.SN
	dev.Manufacturer = base.Manufacturer
	dev.Model = base.Model
	dev.Arch = base.Arch
	dev.ChassisType = base.ChassisType
	dev.Height = base.Height

	raw, _ := protocol.Dispense("raid")
	raid_service := raw.(shared.RaidService)
	////raw, err = protocol.Dispense("oob")
	////oob_service := raw.(shared.OobService)
	//
	dev.Spec, _ = raid_service.RAID()
	data, err := json.MarshalIndent(dev, "", "    ")
	handleErr(err, false)
	fmt.Println("===============Base===============", "\n", string(data))
}

func PostDeviceInfo() {

}