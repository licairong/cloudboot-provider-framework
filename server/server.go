package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/licairong/cloudboot-provider-framework/shared"
	"fmt"
	"log"
	"os/exec"
)

func GetProtocol() (plugin.ClientProtocol, error) {
	pluginClientConfig := &plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Cmd:              exec.Command("./dell-poweredge-r730"),
		Plugins:          map[string]plugin.Plugin{
			"raid": &shared.GRPCRaidPlugin{},
			"oob": &shared.GRPCOobPlugin{},
		},
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	}

	client := plugin.NewClient(pluginClientConfig)
	pluginClientConfig.Reattach = client.ReattachConfig()
	protocol, err := client.Client()
	return protocol, err
}

func main() {
	protocol, err := GetProtocol()
	if err != nil {
		log.Fatalln(err)
	}
	raw, err := protocol.Dispense("raid")
	raid_service := raw.(shared.RaidService)
	raw, err = protocol.Dispense("oob")
	oob_service := raw.(shared.OobService)

	res, _ := raid_service.RAID()
	fmt.Println(res)
	res, _ = raid_service.Clear("0")
	fmt.Println(res)

	res, _ = oob_service.OOB()
	fmt.Println(res)
	res, _ = oob_service.PowerReset()
	fmt.Println(res)
}
