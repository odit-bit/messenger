package chat

import (
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
)

type Message struct {
	ID      string
	Sender  string
	Content string
	Created time.Time
}

func NewMessage(sender, content string) (*Message, error) {
	if sender == "" {
		return nil, fmt.Errorf("sender cannot empty")
	}

	if content == "" {
		return nil, fmt.Errorf("content cannot empty")
	}

	id := ksuid.New()
	return &Message{
		ID:      id.String(),
		Sender:  sender,
		Content: content,
		Created: time.Now(),
	}, nil
}
