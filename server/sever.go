package main

import (

)

func main() {
	// 自检，获取品牌、型号、SN
	SelfCheck()

	// 加载provider
	LoadProvider()

	//// 安装工具
	InstallTools()

	// 发起完整采集
	CollectDeviceInfo()

	// 上报设备信息
	PostDeviceInfo()

	//res, _ := raid_service.RAID()
	//fmt.Println(res)
	//res, _ = raid_service.Clear("0")
	//fmt.Println(res)
	//
	//res, _ = oob_service.OOB()
	//fmt.Println(res)
	//res, _ = oob_service.PowerReset()
	//fmt.Println(res)
}
