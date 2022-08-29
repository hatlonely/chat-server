package service

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/hatlonely/chat-server/api/gen/go/api"
	"github.com/hatlonely/go-kit/logger"
)

type Options struct {
}

func NewChatServiceWithOptions(options *Options) (*ChatService, error) {
	return &ChatService{
		options: options,
		conns:   sync.Map{},
		rpcLog:  logger.NewStdoutJsonLogger(),
	}, nil
}

type ChatService struct {
	api.UnsafeChatServiceServer

	options *Options
	conns   sync.Map
	rpcLog  *logger.Logger
}

func (s *ChatService) setErr(stream api.ChatService_ChatServer, code api.ServerMessage_Err_Code, message string) error {
	res := &api.ServerMessage{
		Type: api.ServerMessage_SMTErr,
		Err: &api.ServerMessage_Err{
			Code:    code,
			Message: message,
		},
	}
	if err := stream.Send(res); err != nil {
		s.rpcLog.Error(err)
		return errors.Wrap(err, "stream.Send failed")
	}
	s.rpcLog.Error(res)

	return errors.New(message)
}

func (s *ChatService) auth(stream api.ChatService_ChatServer) (*api.ClientMessage_Auth, error) {
	message, err := stream.Recv()
	if err != nil {
		return nil, errors.Wrap(err, "stream.Recv failed")
	}
	s.rpcLog.Info(message)

	// 拒绝非授权请求
	if message.Type != api.ClientMessage_CMTAuth {
		return nil, s.setErr(stream, api.ServerMessage_Err_ProtocolMismatch, "协议错误：需要授权信息")
	}
	// 处理授权
	res := &api.ServerMessage{
		Type: api.ServerMessage_SMTAuth,
	}
	if err := stream.Send(res); err != nil {
		s.rpcLog.Error(err)
		return nil, errors.Wrap(err, "stream.Send failed")
	}
	s.rpcLog.Info(res)

	return message.Auth, nil
}

func (s *ChatService) conn(stream api.ChatService_ChatServer, auth *api.ClientMessage_Auth) (string, error) {
	s.conns.Store(auth.Username, stream)

	return "", nil
}

func (s *ChatService) chat(stream api.ChatService_ChatServer) (*api.ClientMessage_Chat, error) {
	message, err := stream.Recv()
	if err != nil {
		return nil, errors.Wrap(err, "stream.Recv failed")
	}
	s.rpcLog.Info(message)

	if message.Type != api.ClientMessage_CMTChat {
		return nil, s.setErr(stream, api.ServerMessage_Err_ProtocolMismatch, "协议错误：需要聊天信息")
	}

	return message.Chat, nil
}

func (s *ChatService) chatLoop(stream api.ChatService_ChatServer, msgChan chan<- *api.ClientMessage_Chat, errChan chan<- error) {
	for {
		message, err := s.chat(stream)
		if err != nil {
			errChan <- err
			break
		}
		msgChan <- message
	}
}

func (s *ChatService) send(stream api.ChatService_ChatServer, username string, message *api.ClientMessage_Chat) error {
	conn, ok := s.conns.Load(message.To)
	if !ok {
		return s.setErr(stream, api.ServerMessage_Err_PersonNotFound, "用户不在线")
	}

	toStream := conn.(api.ChatService_ChatServer)

	res := &api.ServerMessage{
		Type: api.ServerMessage_SMTChat,
		Chat: &api.ServerMessage_Chat{
			From:    username,
			Content: message.Content,
		},
	}
	if err := toStream.Send(res); err != nil {
		s.rpcLog.Error(err)
		return errors.Wrap(err, "stream.Send failed")
	}
	s.rpcLog.Info(res)

	return nil
}

func (s *ChatService) Chat(stream api.ChatService_ChatServer) error {
	auth, err := s.auth(stream)
	if err != nil {
		return errors.WithMessage(err, "auth failed")
	}

	_, err = s.conn(stream, auth)
	if err != nil {
		return errors.WithMessagef(err, "conn failed")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 5)
	msgChan := make(chan *api.ClientMessage_Chat, 5)
	defer close(msgChan)
	defer close(errChan)

	go s.chatLoop(stream, msgChan, errChan)

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			cancel()
			return err
		case msg := <-msgChan:
			if err := s.send(stream, auth.Username, msg); err != nil {
				errChan <- err
			}
		}
	}
}
