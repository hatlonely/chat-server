package service

import (
	"fmt"

	"github.com/hatlonely/chat-server/api/gen/go/api"
)

type Options struct {
}

func NewChatServiceWithOptions(options *Options) (*ChatService, error) {
	return &ChatService{
		options: options,
	}, nil
}

type ChatService struct {
	api.UnsafeChatServiceServer

	options *Options

	chatChannels map[string]*chan api.MessageFromClient
}

func (s *ChatService) Chat(stream api.ChatService_ChatServer) error {
	for {
		message, err := stream.Recv()
		fmt.Println(message, err)
		if err != nil {
			break
		}

		err = stream.Send(&api.MessageFromServer{
			From:    "server",
			Message: "hello client",
		})
		fmt.Println(err)
		if err != nil {
			break
		}
	}
	return nil
}
