package shared

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/licairong/cloudboot-provider-framework/proto"
	"google.golang.org/grpc"
)

type RaidService interface {
	RAID() (string, error)
	Clear(ctrlID string) (string, error)
}

// GRPCHelloPlugin implement plugin.GRPCPlugin
type GRPCRaidPlugin struct {
	plugin.Plugin
	Impl RaidService
}

func (p GRPCRaidPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	proto.RegisterRaidPluginServer(server, GRPCRaidPluginServerWrapper{impl: p.Impl})
	return nil
}

func (p GRPCRaidPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return GRPCRaidPluginClientWrapper{client: proto.NewRaidPluginClient(conn)}, nil
}

type GRPCRaidPluginServerWrapper struct {
	impl RaidService
	proto.UnimplementedRaidPluginServer
}

func (_this GRPCRaidPluginServerWrapper) RAID(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	r, _ := _this.impl.RAID()
	return &proto.Response{
		Result: r,
	}, nil
}

func (_this GRPCRaidPluginServerWrapper) Clear(ctx context.Context, request *proto.ClearRequest) (*proto.Response, error) {
	r, _ := _this.impl.Clear(request.CtrlID)
	return &proto.Response{
		Result: r,
	}, nil
}

// GRPCRaidPluginClientWrapper 作为server 调用插件接口的包装器，
type GRPCRaidPluginClientWrapper struct {
	client proto.RaidPluginClient
}

func (_this GRPCRaidPluginClientWrapper) RAID() (string, error) {
	in := proto.Request{}
	resp, err := _this.client.RAID(context.Background(), &in)
	if err != nil {
		return "", err
	} else {
		return resp.Result, nil
	}
}

func (_this GRPCRaidPluginClientWrapper) Clear(ctrlID string) (string, error) {
	in := proto.ClearRequest{CtrlID: ctrlID}
	resp, err := _this.client.Clear(context.Background(), &in)
	if err != nil {
		return "", err
	} else {
		return resp.Result, nil
	}
}
