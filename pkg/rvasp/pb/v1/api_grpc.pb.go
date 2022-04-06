// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package api

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TRISADemoClient is the client API for TRISADemo service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TRISADemoClient interface {
	LiveUpdates(ctx context.Context, opts ...grpc.CallOption) (TRISADemo_LiveUpdatesClient, error)
}

type tRISADemoClient struct {
	cc grpc.ClientConnInterface
}

func NewTRISADemoClient(cc grpc.ClientConnInterface) TRISADemoClient {
	return &tRISADemoClient{cc}
}

func (c *tRISADemoClient) LiveUpdates(ctx context.Context, opts ...grpc.CallOption) (TRISADemo_LiveUpdatesClient, error) {
	stream, err := c.cc.NewStream(ctx, &TRISADemo_ServiceDesc.Streams[0], "/rvasp.v1.TRISADemo/LiveUpdates", opts...)
	if err != nil {
		return nil, err
	}
	x := &tRISADemoLiveUpdatesClient{stream}
	return x, nil
}

type TRISADemo_LiveUpdatesClient interface {
	Send(*Command) error
	Recv() (*Message, error)
	grpc.ClientStream
}

type tRISADemoLiveUpdatesClient struct {
	grpc.ClientStream
}

func (x *tRISADemoLiveUpdatesClient) Send(m *Command) error {
	return x.ClientStream.SendMsg(m)
}

func (x *tRISADemoLiveUpdatesClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TRISADemoServer is the server API for TRISADemo service.
// All implementations must embed UnimplementedTRISADemoServer
// for forward compatibility
type TRISADemoServer interface {
	LiveUpdates(TRISADemo_LiveUpdatesServer) error
	mustEmbedUnimplementedTRISADemoServer()
}

// UnimplementedTRISADemoServer must be embedded to have forward compatible implementations.
type UnimplementedTRISADemoServer struct {
}

func (UnimplementedTRISADemoServer) LiveUpdates(TRISADemo_LiveUpdatesServer) error {
	return status.Errorf(codes.Unimplemented, "method LiveUpdates not implemented")
}
func (UnimplementedTRISADemoServer) mustEmbedUnimplementedTRISADemoServer() {}

// UnsafeTRISADemoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TRISADemoServer will
// result in compilation errors.
type UnsafeTRISADemoServer interface {
	mustEmbedUnimplementedTRISADemoServer()
}

func RegisterTRISADemoServer(s grpc.ServiceRegistrar, srv TRISADemoServer) {
	s.RegisterService(&TRISADemo_ServiceDesc, srv)
}

func _TRISADemo_LiveUpdates_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TRISADemoServer).LiveUpdates(&tRISADemoLiveUpdatesServer{stream})
}

type TRISADemo_LiveUpdatesServer interface {
	Send(*Message) error
	Recv() (*Command, error)
	grpc.ServerStream
}

type tRISADemoLiveUpdatesServer struct {
	grpc.ServerStream
}

func (x *tRISADemoLiveUpdatesServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func (x *tRISADemoLiveUpdatesServer) Recv() (*Command, error) {
	m := new(Command)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TRISADemo_ServiceDesc is the grpc.ServiceDesc for TRISADemo service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TRISADemo_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rvasp.v1.TRISADemo",
	HandlerType: (*TRISADemoServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "LiveUpdates",
			Handler:       _TRISADemo_LiveUpdates_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "rvasp/v1/api.proto",
}

// TRISAIntegrationClient is the client API for TRISAIntegration service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TRISAIntegrationClient interface {
	Transfer(ctx context.Context, in *TransferRequest, opts ...grpc.CallOption) (*TransferReply, error)
	AccountStatus(ctx context.Context, in *AccountRequest, opts ...grpc.CallOption) (*AccountReply, error)
}

type tRISAIntegrationClient struct {
	cc grpc.ClientConnInterface
}

func NewTRISAIntegrationClient(cc grpc.ClientConnInterface) TRISAIntegrationClient {
	return &tRISAIntegrationClient{cc}
}

func (c *tRISAIntegrationClient) Transfer(ctx context.Context, in *TransferRequest, opts ...grpc.CallOption) (*TransferReply, error) {
	out := new(TransferReply)
	err := c.cc.Invoke(ctx, "/rvasp.v1.TRISAIntegration/Transfer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tRISAIntegrationClient) AccountStatus(ctx context.Context, in *AccountRequest, opts ...grpc.CallOption) (*AccountReply, error) {
	out := new(AccountReply)
	err := c.cc.Invoke(ctx, "/rvasp.v1.TRISAIntegration/AccountStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TRISAIntegrationServer is the server API for TRISAIntegration service.
// All implementations must embed UnimplementedTRISAIntegrationServer
// for forward compatibility
type TRISAIntegrationServer interface {
	Transfer(context.Context, *TransferRequest) (*TransferReply, error)
	AccountStatus(context.Context, *AccountRequest) (*AccountReply, error)
	mustEmbedUnimplementedTRISAIntegrationServer()
}

// UnimplementedTRISAIntegrationServer must be embedded to have forward compatible implementations.
type UnimplementedTRISAIntegrationServer struct {
}

func (UnimplementedTRISAIntegrationServer) Transfer(context.Context, *TransferRequest) (*TransferReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transfer not implemented")
}
func (UnimplementedTRISAIntegrationServer) AccountStatus(context.Context, *AccountRequest) (*AccountReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AccountStatus not implemented")
}
func (UnimplementedTRISAIntegrationServer) mustEmbedUnimplementedTRISAIntegrationServer() {}

// UnsafeTRISAIntegrationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TRISAIntegrationServer will
// result in compilation errors.
type UnsafeTRISAIntegrationServer interface {
	mustEmbedUnimplementedTRISAIntegrationServer()
}

func RegisterTRISAIntegrationServer(s grpc.ServiceRegistrar, srv TRISAIntegrationServer) {
	s.RegisterService(&TRISAIntegration_ServiceDesc, srv)
}

func _TRISAIntegration_Transfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransferRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TRISAIntegrationServer).Transfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rvasp.v1.TRISAIntegration/Transfer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TRISAIntegrationServer).Transfer(ctx, req.(*TransferRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TRISAIntegration_AccountStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TRISAIntegrationServer).AccountStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rvasp.v1.TRISAIntegration/AccountStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TRISAIntegrationServer).AccountStatus(ctx, req.(*AccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TRISAIntegration_ServiceDesc is the grpc.ServiceDesc for TRISAIntegration service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TRISAIntegration_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rvasp.v1.TRISAIntegration",
	HandlerType: (*TRISAIntegrationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Transfer",
			Handler:    _TRISAIntegration_Transfer_Handler,
		},
		{
			MethodName: "AccountStatus",
			Handler:    _TRISAIntegration_AccountStatus_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rvasp/v1/api.proto",
}
