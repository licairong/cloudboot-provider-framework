package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/licairong/hardware-plugin/shared"
	"github.com/licairong/raid-avago"
	"github.com/licairong/oob-dell"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"RaidPlugin": &shared.GRPCRaidPlugin{Impl: &raid.RaidPlugin{}},
			"OobPlugin": &shared.GRPCOobPlugin{Impl: &oob.OobPlugin{}},
			//"BiosPlugin": &shared.GRPCBiosPlugin{Impl: &bios.BiosPlugin{}},
			//"FirmwarePlugin": &shared.GRPCFirmwarePlugin{Impl: &firmware.FirmwarePlugin{}},
		},
		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
