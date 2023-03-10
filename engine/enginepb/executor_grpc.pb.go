// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package enginepb

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

// ExecutorServiceClient is the client API for ExecutorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ExecutorServiceClient interface {
	PreDispatchTask(ctx context.Context, in *PreDispatchTaskRequest, opts ...grpc.CallOption) (*PreDispatchTaskResponse, error)
	ConfirmDispatchTask(ctx context.Context, in *ConfirmDispatchTaskRequest, opts ...grpc.CallOption) (*ConfirmDispatchTaskResponse, error)
}

type executorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewExecutorServiceClient(cc grpc.ClientConnInterface) ExecutorServiceClient {
	return &executorServiceClient{cc}
}

func (c *executorServiceClient) PreDispatchTask(ctx context.Context, in *PreDispatchTaskRequest, opts ...grpc.CallOption) (*PreDispatchTaskResponse, error) {
	out := new(PreDispatchTaskResponse)
	err := c.cc.Invoke(ctx, "/enginepb.ExecutorService/PreDispatchTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *executorServiceClient) ConfirmDispatchTask(ctx context.Context, in *ConfirmDispatchTaskRequest, opts ...grpc.CallOption) (*ConfirmDispatchTaskResponse, error) {
	out := new(ConfirmDispatchTaskResponse)
	err := c.cc.Invoke(ctx, "/enginepb.ExecutorService/ConfirmDispatchTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExecutorServiceServer is the server API for ExecutorService service.
// All implementations should embed UnimplementedExecutorServiceServer
// for forward compatibility
type ExecutorServiceServer interface {
	PreDispatchTask(context.Context, *PreDispatchTaskRequest) (*PreDispatchTaskResponse, error)
	ConfirmDispatchTask(context.Context, *ConfirmDispatchTaskRequest) (*ConfirmDispatchTaskResponse, error)
}

// UnimplementedExecutorServiceServer should be embedded to have forward compatible implementations.
type UnimplementedExecutorServiceServer struct {
}

func (UnimplementedExecutorServiceServer) PreDispatchTask(context.Context, *PreDispatchTaskRequest) (*PreDispatchTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreDispatchTask not implemented")
}
func (UnimplementedExecutorServiceServer) ConfirmDispatchTask(context.Context, *ConfirmDispatchTaskRequest) (*ConfirmDispatchTaskResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmDispatchTask not implemented")
}

// UnsafeExecutorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ExecutorServiceServer will
// result in compilation errors.
type UnsafeExecutorServiceServer interface {
	mustEmbedUnimplementedExecutorServiceServer()
}

func RegisterExecutorServiceServer(s grpc.ServiceRegistrar, srv ExecutorServiceServer) {
	s.RegisterService(&ExecutorService_ServiceDesc, srv)
}

func _ExecutorService_PreDispatchTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PreDispatchTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorServiceServer).PreDispatchTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/enginepb.ExecutorService/PreDispatchTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorServiceServer).PreDispatchTask(ctx, req.(*PreDispatchTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ExecutorService_ConfirmDispatchTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfirmDispatchTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ExecutorServiceServer).ConfirmDispatchTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/enginepb.ExecutorService/ConfirmDispatchTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ExecutorServiceServer).ConfirmDispatchTask(ctx, req.(*ConfirmDispatchTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ExecutorService_ServiceDesc is the grpc.ServiceDesc for ExecutorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ExecutorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "enginepb.ExecutorService",
	HandlerType: (*ExecutorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PreDispatchTask",
			Handler:    _ExecutorService_PreDispatchTask_Handler,
		},
		{
			MethodName: "ConfirmDispatchTask",
			Handler:    _ExecutorService_ConfirmDispatchTask_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "engine/proto/executor.proto",
}

// BrokerServiceClient is the client API for BrokerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BrokerServiceClient interface {
	RemoveResource(ctx context.Context, in *RemoveLocalResourceRequest, opts ...grpc.CallOption) (*RemoveLocalResourceResponse, error)
}

type brokerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBrokerServiceClient(cc grpc.ClientConnInterface) BrokerServiceClient {
	return &brokerServiceClient{cc}
}

func (c *brokerServiceClient) RemoveResource(ctx context.Context, in *RemoveLocalResourceRequest, opts ...grpc.CallOption) (*RemoveLocalResourceResponse, error) {
	out := new(RemoveLocalResourceResponse)
	err := c.cc.Invoke(ctx, "/enginepb.BrokerService/RemoveResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BrokerServiceServer is the server API for BrokerService service.
// All implementations should embed UnimplementedBrokerServiceServer
// for forward compatibility
type BrokerServiceServer interface {
	RemoveResource(context.Context, *RemoveLocalResourceRequest) (*RemoveLocalResourceResponse, error)
}

// UnimplementedBrokerServiceServer should be embedded to have forward compatible implementations.
type UnimplementedBrokerServiceServer struct {
}

func (UnimplementedBrokerServiceServer) RemoveResource(context.Context, *RemoveLocalResourceRequest) (*RemoveLocalResourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveResource not implemented")
}

// UnsafeBrokerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BrokerServiceServer will
// result in compilation errors.
type UnsafeBrokerServiceServer interface {
	mustEmbedUnimplementedBrokerServiceServer()
}

func RegisterBrokerServiceServer(s grpc.ServiceRegistrar, srv BrokerServiceServer) {
	s.RegisterService(&BrokerService_ServiceDesc, srv)
}

func _BrokerService_RemoveResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveLocalResourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BrokerServiceServer).RemoveResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/enginepb.BrokerService/RemoveResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BrokerServiceServer).RemoveResource(ctx, req.(*RemoveLocalResourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BrokerService_ServiceDesc is the grpc.ServiceDesc for BrokerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BrokerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "enginepb.BrokerService",
	HandlerType: (*BrokerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RemoveResource",
			Handler:    _BrokerService_RemoveResource_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "engine/proto/executor.proto",
}
