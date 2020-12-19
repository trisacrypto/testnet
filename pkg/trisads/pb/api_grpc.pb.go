// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// TRISADirectoryClient is the client API for TRISADirectory service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TRISADirectoryClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error)
	Lookup(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupReply, error)
	Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchReply, error)
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusReply, error)
	VerifyEmail(ctx context.Context, in *VerifyEmailRequest, opts ...grpc.CallOption) (*VerifyEmailReply, error)
}

type tRISADirectoryClient struct {
	cc grpc.ClientConnInterface
}

func NewTRISADirectoryClient(cc grpc.ClientConnInterface) TRISADirectoryClient {
	return &tRISADirectoryClient{cc}
}

func (c *tRISADirectoryClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error) {
	out := new(RegisterReply)
	err := c.cc.Invoke(ctx, "/trisads.TRISADirectory/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tRISADirectoryClient) Lookup(ctx context.Context, in *LookupRequest, opts ...grpc.CallOption) (*LookupReply, error) {
	out := new(LookupReply)
	err := c.cc.Invoke(ctx, "/trisads.TRISADirectory/Lookup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tRISADirectoryClient) Search(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchReply, error) {
	out := new(SearchReply)
	err := c.cc.Invoke(ctx, "/trisads.TRISADirectory/Search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tRISADirectoryClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusReply, error) {
	out := new(StatusReply)
	err := c.cc.Invoke(ctx, "/trisads.TRISADirectory/Status", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tRISADirectoryClient) VerifyEmail(ctx context.Context, in *VerifyEmailRequest, opts ...grpc.CallOption) (*VerifyEmailReply, error) {
	out := new(VerifyEmailReply)
	err := c.cc.Invoke(ctx, "/trisads.TRISADirectory/VerifyEmail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TRISADirectoryServer is the server API for TRISADirectory service.
// All implementations must embed UnimplementedTRISADirectoryServer
// for forward compatibility
type TRISADirectoryServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterReply, error)
	Lookup(context.Context, *LookupRequest) (*LookupReply, error)
	Search(context.Context, *SearchRequest) (*SearchReply, error)
	Status(context.Context, *StatusRequest) (*StatusReply, error)
	VerifyEmail(context.Context, *VerifyEmailRequest) (*VerifyEmailReply, error)
	mustEmbedUnimplementedTRISADirectoryServer()
}

// UnimplementedTRISADirectoryServer must be embedded to have forward compatible implementations.
type UnimplementedTRISADirectoryServer struct {
}

func (UnimplementedTRISADirectoryServer) Register(context.Context, *RegisterRequest) (*RegisterReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedTRISADirectoryServer) Lookup(context.Context, *LookupRequest) (*LookupReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Lookup not implemented")
}
func (UnimplementedTRISADirectoryServer) Search(context.Context, *SearchRequest) (*SearchReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}
func (UnimplementedTRISADirectoryServer) Status(context.Context, *StatusRequest) (*StatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (UnimplementedTRISADirectoryServer) VerifyEmail(context.Context, *VerifyEmailRequest) (*VerifyEmailReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VerifyEmail not implemented")
}
func (UnimplementedTRISADirectoryServer) mustEmbedUnimplementedTRISADirectoryServer() {}

// UnsafeTRISADirectoryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TRISADirectoryServer will
// result in compilation errors.
type UnsafeTRISADirectoryServer interface {
	mustEmbedUnimplementedTRISADirectoryServer()
}

func RegisterTRISADirectoryServer(s grpc.ServiceRegistrar, srv TRISADirectoryServer) {
	s.RegisterService(&_TRISADirectory_serviceDesc, srv)
}

func _TRISADirectory_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TRISADirectoryServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trisads.TRISADirectory/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TRISADirectoryServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TRISADirectory_Lookup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TRISADirectoryServer).Lookup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trisads.TRISADirectory/Lookup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TRISADirectoryServer).Lookup(ctx, req.(*LookupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TRISADirectory_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TRISADirectoryServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trisads.TRISADirectory/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TRISADirectoryServer).Search(ctx, req.(*SearchRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TRISADirectory_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TRISADirectoryServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trisads.TRISADirectory/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TRISADirectoryServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TRISADirectory_VerifyEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerifyEmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TRISADirectoryServer).VerifyEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/trisads.TRISADirectory/VerifyEmail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TRISADirectoryServer).VerifyEmail(ctx, req.(*VerifyEmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TRISADirectory_serviceDesc = grpc.ServiceDesc{
	ServiceName: "trisads.TRISADirectory",
	HandlerType: (*TRISADirectoryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _TRISADirectory_Register_Handler,
		},
		{
			MethodName: "Lookup",
			Handler:    _TRISADirectory_Lookup_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _TRISADirectory_Search_Handler,
		},
		{
			MethodName: "Status",
			Handler:    _TRISADirectory_Status_Handler,
		},
		{
			MethodName: "VerifyEmail",
			Handler:    _TRISADirectory_VerifyEmail_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "trisads/api.proto",
}
