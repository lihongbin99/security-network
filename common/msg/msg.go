package msg

import "fmt"

type Message interface {
	GetMessageType() uint32
}

const (
	_ uint32 = iota

	PingMessageType
	PoneMessageType
)

func NewMessage(messageType uint32) (message Message, err error) {
	switch messageType {
	case PingMessageType:
		message = &PingMessage{}
	case PoneMessageType:
		message = &PoneMessage{}
	default:
		err = fmt.Errorf("no find message type: %d", messageType)
	}
	return
}
