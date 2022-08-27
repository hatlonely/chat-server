package service

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/hatlonely/chat-server/api/gen/go/api"
)

type Options struct {
}

func NewChatServiceWithOptions(options *Options) (*ChatService, error) {
	return &ChatService{
		options:      options,
		chatChannels: sync.Map{},
	}, nil
}

type ChatService struct {
	api.UnsafeChatServiceServer

	options *Options

	// chatChannels map[string]*chan api.ClientMessage
	chatChannels sync.Map
}

func (s *ChatService) Chat(stream api.ChatService_ChatServer) error {
	auth, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "stream.Recv failed")
	}
	// 拒绝非授权请求
	if auth.Type != api.ClientMessage_CMTAuth {
		if err := stream.Send(&api.ServerMessage{
			Type: api.ServerMessage_SMTErr,
			Err: &api.ServerMessage_Err{
				Code:    api.ServerMessage_Err_ProtocolMismatch,
				Message: "无授权信息",
			},
		}); err != nil {
			return errors.Wrap(err, "stream.Send failed")
		}

		return nil
	}
	// 处理授权
	if err := stream.Send(&api.ServerMessage{
		Type: api.ServerMessage_SMTAuth,
	}); err != nil {
		return errors.Wrap(err, "stream.Send failed")
	}
	username := auth.Auth.Username
	chatChannel := make(chan *api.ServerMessage_Chat, 5)
	s.chatChannels.Store(username, chatChannel)
	defer s.chatChannels.Delete(username)
	defer close(chatChannel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	// 从客户端接收消息
	go func() {
		for {
			message, err := stream.Recv()
			if err != nil {
				errChan <- errors.Wrap(err, "stream.Recv failed")
				return
			}

			if message.Type != api.ClientMessage_CMTChat {
				if err := stream.Send(&api.ServerMessage{
					Type: api.ServerMessage_SMTErr,
					Err: &api.ServerMessage_Err{
						Code:    api.ServerMessage_Err_ProtocolMismatch,
						Message: "无授权信息",
					},
				}); err != nil {
					errChan <- errors.Wrap(err, "stream.Send failed")
				}
				return
			}

			ch, ok := s.chatChannels.Load(message.Chat.To)
			if !ok {
				if err := stream.Send(&api.ServerMessage{
					Type: api.ServerMessage_SMTErr,
					Err: &api.ServerMessage_Err{
						Code:    api.ServerMessage_Err_PersonNotFound,
						Message: "用户不在线",
					},
				}); err != nil {
					errChan <- errors.Wrap(err, "stream.Send failed")
				}
				return
			}

			ch.(chan *api.ServerMessage_Chat) <- &api.ServerMessage_Chat{
				From:    username,
				Content: message.Chat.Content,
			}
		}
	}()

	// 取消息发送给客户端
	go func() {
		for message := range chatChannel {
			if err := stream.Send(&api.ServerMessage{
				Type: api.ServerMessage_SMTChat,
				Chat: message,
			}); err != nil {
				errChan <- errors.Wrap(err, "stream.Send failed")
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			return err
		}
	}
}
