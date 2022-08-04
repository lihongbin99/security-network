package main

import (
	"flag"
	"fmt"
	"net"
	"security-network/common/config"
	"security-network/common/io"
	"security-network/common/logger"
	"security-network/common/msg"
	"time"
)

var (
	log = logger.NewLog("Client")
)

func init() {
	flag.Parse()

	// 默认的配置文件路径
	if config.File == "" {
		config.File = "config/client.ini"
	}

	config.Init()
	config.ClientInit()
}

func main() {
	interval := 1
	for {
		success := start()
		if success {
			interval = 1
		}
		log.Debug("sleep:", interval)
		time.Sleep(time.Duration(interval) * time.Second)
		interval = interval * 2
		if interval > 60 {
			interval = 60
		}
	}
}

func start() (success bool) {
	serverAddr := fmt.Sprintf("%s:%d", config.ServerIp, config.ServerPort)
	conn, success := connectServer(serverAddr)
	if !success {
		return success
	}
	defer func() {
		_ = conn.Close()
		log.Info("close success:", serverAddr)
	}()
	log.Debug("connect server success:", serverAddr)

	serverConn := io.NewTCP(conn)
	serverConn.ClientInit()
	log.Info("new server:", serverAddr)

	// 处理读取请求
	readChan := make(chan io.Message, 8)
	go func(tcp *io.TCP, readChan chan io.Message) {
		defer close(readChan)
		for {
			message := tcp.ReadMessage(time.Time{})
			readChan <- message
			if message.Err != nil {
				break
			}
		}
	}(serverConn, readChan)

	// 心跳
	pingTicker := time.NewTicker(time.Duration(serverConn.Interval+10) * time.Second)
	defer pingTicker.Stop()
	lastPongTime := time.Now()
	lastPingTime := time.Now()

	var err error = nil
	for err == nil {
		select {
		case <-pingTicker.C:
			if lastPongTime.Before(lastPingTime) {
				err = fmt.Errorf("ping timeout")
			}
			lastPingTime = time.Now()
		case message := <-readChan:
			if message.Err != nil {
				err = message.Err
				break
			}
			switch m := message.Message.(type) {
			case *msg.PingMessage:
				lastPongTime = time.Now()
				log.Trace("receiver PingMessage", m.Date)
				err = serverConn.WriteMessage(&msg.PoneMessage{Date: time.Now()})
			}
		}
	}

	log.Info("exit:", err)
	return true
}

func connectServer(serverAddr string) (*net.TCPConn, bool) {
	addr, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Error("resolve tcp addr error", err)
		return nil, false
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Error("dial tcp addr error", err)
		return nil, false
	}
	return conn, true
}
