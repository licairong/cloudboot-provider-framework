package shared

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/licairong/cloudboot-provider-framework/proto"
	"google.golang.org/grpc"
)

type OobService interface {
	OOB() (string, error)
	PowerReset() (string, error)
}

type GRPCOobPlugin struct {
	plugin.Plugin
	Impl OobService
}

func (p GRPCOobPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	proto.RegisterOobPluginServer(server, GRPCOobPluginServerWrapper{impl: p.Impl})
	return nil
}

func (p GRPCOobPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return GRPCOobPluginClientWrapper{client: proto.NewOobPluginClient(conn)}, nil
}

type GRPCOobPluginServerWrapper struct {
	impl OobService
	proto.UnimplementedOobPluginServer
}

func (_this GRPCOobPluginServerWrapper) OOB(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	r, _ := _this.impl.OOB()
	return &proto.Response{
		Result: r,
	}, nil
}

func (_this GRPCOobPluginServerWrapper) PowerReset(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	r, _ := _this.impl.PowerReset()
	return &proto.Response{
		Result: r,
	}, nil
}

type GRPCOobPluginClientWrapper struct {
	client proto.OobPluginClient
}

func (_this GRPCOobPluginClientWrapper) OOB() (string, error) {
	in := proto.Request{}
	resp, err := _this.client.OOB(context.Background(), &in)
	if err != nil {
		return "", err
	} else {
		return resp.Result, nil
	}
}

func (_this GRPCOobPluginClientWrapper) PowerReset() (string, error) {
	in := proto.Request{}
	resp, err := _this.client.PowerReset(context.Background(), &in)
	if err != nil {
		return "", err
	} else {
		return resp.Result, nil
	}
}
