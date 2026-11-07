package message

import "errors"

type ID string

type Message struct {
	id      ID
	content string
}

func NewMessage(id ID, content string) (*Message, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}
	return &Message{id: id, content: content}, nil
}

func (m *Message) ID() ID          { return m.id }
func (m *Message) Content() string { return m.content }
