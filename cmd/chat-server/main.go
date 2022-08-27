package main

import (
	"fmt"
	"net"

	"github.com/hatlonely/chat-server/api/gen/go/api"
	"github.com/hatlonely/chat-server/internal/service"

	"github.com/hatlonely/go-kit/refx"
	"google.golang.org/grpc"
)

func main() {
	svc, err := service.NewChatServiceWithOptions(&service.Options{})
	refx.Must(err)

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 6080))
	refx.Must(err)

	grpcServer := grpc.NewServer()
	api.RegisterChatServiceServer(grpcServer, svc)
	refx.Must(grpcServer.Serve(listener))
}
