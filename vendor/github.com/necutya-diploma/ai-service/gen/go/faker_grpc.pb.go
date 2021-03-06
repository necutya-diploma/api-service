// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package _go

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

// AIClient is the client API for AI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AIClient interface {
	CheckMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*MessageResponse, error)
}

type aIClient struct {
	cc grpc.ClientConnInterface
}

func NewAIClient(cc grpc.ClientConnInterface) AIClient {
	return &aIClient{cc}
}

func (c *aIClient) CheckMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*MessageResponse, error) {
	out := new(MessageResponse)
	err := c.cc.Invoke(ctx, "/faker.AI/CheckMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AIServer is the server API for AI service.
// All implementations must embed UnimplementedAIServer
// for forward compatibility
type AIServer interface {
	CheckMessage(context.Context, *Message) (*MessageResponse, error)
	mustEmbedUnimplementedAIServer()
}

// UnimplementedAIServer must be embedded to have forward compatible implementations.
type UnimplementedAIServer struct {
}

func (UnimplementedAIServer) CheckMessage(context.Context, *Message) (*MessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckMessage not implemented")
}
func (UnimplementedAIServer) mustEmbedUnimplementedAIServer() {}

// UnsafeAIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AIServer will
// result in compilation errors.
type UnsafeAIServer interface {
	mustEmbedUnimplementedAIServer()
}

func RegisterAIServer(s grpc.ServiceRegistrar, srv AIServer) {
	s.RegisterService(&AI_ServiceDesc, srv)
}

func _AI_CheckMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AIServer).CheckMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/faker.AI/CheckMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AIServer).CheckMessage(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

// AI_ServiceDesc is the grpc.ServiceDesc for AI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "faker.AI",
	HandlerType: (*AIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckMessage",
			Handler:    _AI_CheckMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/faker.proto",
}
