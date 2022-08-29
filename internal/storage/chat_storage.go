package storage

type ChatStorage interface {
	PutMessage(from string, to string, content string) error
	GetMessageByUser(from string, seq int64) []*ChatMessage
}
