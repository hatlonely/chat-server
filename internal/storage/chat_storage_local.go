package storage

import (
	"sync"
	"time"
)

type LocalChatStorageOptions struct {
}

func NewLocalChatStorageWithOptions() *LocalChatStorage {
	return &LocalChatStorage{
		userMessagesMap: map[string]*ChatMessages{},
	}
}

type LocalChatStorage struct {
	userMessagesMap map[string]*ChatMessages
	mutex           sync.RWMutex
}

type ChatMessage struct {
	Seq       int64
	Timestamp time.Time
	From      string
	To        string
	Content   string
}

type ChatMessages struct {
	seq      int64
	messages []*ChatMessage
	mutex    sync.RWMutex
}

func (m *ChatMessages) Append(from string, to string, content string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.seq += 1
	m.messages = append(m.messages, &ChatMessage{
		Timestamp: time.Now(),
		Seq:       m.seq,
		From:      from,
		To:        to,
		Content:   content,
	})
}

func (m *ChatMessages) Lookup(seq int64) []*ChatMessage {
	var messages []*ChatMessage
	m.mutex.RLock()
	for _, message := range m.messages {
		if message.Seq < seq {
			continue
		}
		messages = append(messages, message)
	}
	m.mutex.RUnlock()
	return messages
}

func (s *LocalChatStorage) PutMessage(from string, to string, content string) error {
	s.putOneMessage(from, from, to, content)
	s.putOneMessage(to, from, to, content)
	return nil
}

func (s *LocalChatStorage) GetMessageByUser(from string, seq int64) []*ChatMessage {
	s.mutex.RLock()
	messages, ok := s.userMessagesMap[from]
	s.mutex.RUnlock()
	if !ok {
		return nil
	}

	return messages.Lookup(seq)
}

func (s *LocalChatStorage) putOneMessage(key string, from string, to string, content string) {
	s.mutex.RLock()
	messages, ok := s.userMessagesMap[key]
	s.mutex.RUnlock()
	if !ok {
		s.mutex.Lock()
		messages = &ChatMessages{}
		s.userMessagesMap[key] = messages
		s.mutex.Unlock()
	}
	messages.Append(from, to, content)
}
