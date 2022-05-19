// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package gen

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

// MailerClient is the client API for Mailer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MailerClient interface {
	SendEmail(ctx context.Context, in *EmailMessage, opts ...grpc.CallOption) (*EmptyResponse, error)
}

type mailerClient struct {
	cc grpc.ClientConnInterface
}

func NewMailerClient(cc grpc.ClientConnInterface) MailerClient {
	return &mailerClient{cc}
}

func (c *mailerClient) SendEmail(ctx context.Context, in *EmailMessage, opts ...grpc.CallOption) (*EmptyResponse, error) {
	out := new(EmptyResponse)
	err := c.cc.Invoke(ctx, "/mailer.Mailer/SendEmail", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MailerServer is the server API for Mailer service.
// All implementations must embed UnimplementedMailerServer
// for forward compatibility
type MailerServer interface {
	SendEmail(context.Context, *EmailMessage) (*EmptyResponse, error)
	mustEmbedUnimplementedMailerServer()
}

// UnimplementedMailerServer must be embedded to have forward compatible implementations.
type UnimplementedMailerServer struct {
}

func (UnimplementedMailerServer) SendEmail(context.Context, *EmailMessage) (*EmptyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendEmail not implemented")
}
func (UnimplementedMailerServer) mustEmbedUnimplementedMailerServer() {}

// UnsafeMailerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MailerServer will
// result in compilation errors.
type UnsafeMailerServer interface {
	mustEmbedUnimplementedMailerServer()
}

func RegisterMailerServer(s grpc.ServiceRegistrar, srv MailerServer) {
	s.RegisterService(&Mailer_ServiceDesc, srv)
}

func _Mailer_SendEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmailMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MailerServer).SendEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mailer.Mailer/SendEmail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MailerServer).SendEmail(ctx, req.(*EmailMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// Mailer_ServiceDesc is the grpc.ServiceDesc for Mailer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Mailer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mailer.Mailer",
	HandlerType: (*MailerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendEmail",
			Handler:    _Mailer_SendEmail_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/mailer.proto",
}
