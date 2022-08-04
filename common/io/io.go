package io

import (
	"security-network/common/logger"
	"security-network/common/msg"
)

var (
	log = logger.NewLog("IO")
)

type Message struct {
	Message msg.Message
	Err     error
}
