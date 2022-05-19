// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package thprotos

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

// TempHumSvcClient is the client API for TempHumSvc service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TempHumSvcClient interface {
	// Read Temperature and Humidity from sensor
	ReadTempHum(ctx context.Context, in *ReadTempHumRequest, opts ...grpc.CallOption) (*ReadTempHumReply, error)
}

type tempHumSvcClient struct {
	cc grpc.ClientConnInterface
}

func NewTempHumSvcClient(cc grpc.ClientConnInterface) TempHumSvcClient {
	return &tempHumSvcClient{cc}
}

func (c *tempHumSvcClient) ReadTempHum(ctx context.Context, in *ReadTempHumRequest, opts ...grpc.CallOption) (*ReadTempHumReply, error) {
	out := new(ReadTempHumReply)
	err := c.cc.Invoke(ctx, "/thgrpc.TempHumSvc/ReadTempHum", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TempHumSvcServer is the server API for TempHumSvc service.
// All implementations must embed UnimplementedTempHumSvcServer
// for forward compatibility
type TempHumSvcServer interface {
	// Read Temperature and Humidity from sensor
	ReadTempHum(context.Context, *ReadTempHumRequest) (*ReadTempHumReply, error)
	mustEmbedUnimplementedTempHumSvcServer()
}

// UnimplementedTempHumSvcServer must be embedded to have forward compatible implementations.
type UnimplementedTempHumSvcServer struct {
}

func (UnimplementedTempHumSvcServer) ReadTempHum(context.Context, *ReadTempHumRequest) (*ReadTempHumReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadTempHum not implemented")
}
func (UnimplementedTempHumSvcServer) mustEmbedUnimplementedTempHumSvcServer() {}

// UnsafeTempHumSvcServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TempHumSvcServer will
// result in compilation errors.
type UnsafeTempHumSvcServer interface {
	mustEmbedUnimplementedTempHumSvcServer()
}

func RegisterTempHumSvcServer(s grpc.ServiceRegistrar, srv TempHumSvcServer) {
	s.RegisterService(&TempHumSvc_ServiceDesc, srv)
}

func _TempHumSvc_ReadTempHum_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadTempHumRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TempHumSvcServer).ReadTempHum(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/thgrpc.TempHumSvc/ReadTempHum",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TempHumSvcServer).ReadTempHum(ctx, req.(*ReadTempHumRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TempHumSvc_ServiceDesc is the grpc.ServiceDesc for TempHumSvc service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TempHumSvc_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "thgrpc.TempHumSvc",
	HandlerType: (*TempHumSvcServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReadTempHum",
			Handler:    _TempHumSvc_ReadTempHum_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "thsock.proto",
}
