// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: api/chat-server.proto

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

// ChatServiceClient is the client API for ChatService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ChatServiceClient interface {
	Chat(ctx context.Context, opts ...grpc.CallOption) (ChatService_ChatClient, error)
}

type chatServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewChatServiceClient(cc grpc.ClientConnInterface) ChatServiceClient {
	return &chatServiceClient{cc}
}

func (c *chatServiceClient) Chat(ctx context.Context, opts ...grpc.CallOption) (ChatService_ChatClient, error) {
	stream, err := c.cc.NewStream(ctx, &ChatService_ServiceDesc.Streams[0], "/api.ChatService/Chat", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatServiceChatClient{stream}
	return x, nil
}

type ChatService_ChatClient interface {
	Send(*MessageToSend) error
	Recv() (*MessageToRecv, error)
	grpc.ClientStream
}

type chatServiceChatClient struct {
	grpc.ClientStream
}

func (x *chatServiceChatClient) Send(m *MessageToSend) error {
	return x.ClientStream.SendMsg(m)
}

func (x *chatServiceChatClient) Recv() (*MessageToRecv, error) {
	m := new(MessageToRecv)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChatServiceServer is the server API for ChatService service.
// All implementations must embed UnimplementedChatServiceServer
// for forward compatibility
type ChatServiceServer interface {
	Chat(ChatService_ChatServer) error
	mustEmbedUnimplementedChatServiceServer()
}

// UnimplementedChatServiceServer must be embedded to have forward compatible implementations.
type UnimplementedChatServiceServer struct {
}

func (UnimplementedChatServiceServer) Chat(ChatService_ChatServer) error {
	return status.Errorf(codes.Unimplemented, "method Chat not implemented")
}
func (UnimplementedChatServiceServer) mustEmbedUnimplementedChatServiceServer() {}

// UnsafeChatServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ChatServiceServer will
// result in compilation errors.
type UnsafeChatServiceServer interface {
	mustEmbedUnimplementedChatServiceServer()
}

func RegisterChatServiceServer(s grpc.ServiceRegistrar, srv ChatServiceServer) {
	s.RegisterService(&ChatService_ServiceDesc, srv)
}

func _ChatService_Chat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ChatServiceServer).Chat(&chatServiceChatServer{stream})
}

type ChatService_ChatServer interface {
	Send(*MessageToRecv) error
	Recv() (*MessageToSend, error)
	grpc.ServerStream
}

type chatServiceChatServer struct {
	grpc.ServerStream
}

func (x *chatServiceChatServer) Send(m *MessageToRecv) error {
	return x.ServerStream.SendMsg(m)
}

func (x *chatServiceChatServer) Recv() (*MessageToSend, error) {
	m := new(MessageToSend)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChatService_ServiceDesc is the grpc.ServiceDesc for ChatService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ChatService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.ChatService",
	HandlerType: (*ChatServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Chat",
			Handler:       _ChatService_Chat_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "api/chat-server.proto",
}
