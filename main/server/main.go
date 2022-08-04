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

type Client struct {
	ip   string
	name string
}

var (
	log = logger.NewLog("Server")

	clients = make(map[int]Client)
)

func init() {
	flag.Parse()

	// 默认的配置文件路径
	if config.File == "" {
		config.File = "config/server.ini"
	}

	config.Init()
	config.ServerInit()
}

func main() {
	var err error = nil
	listenAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", config.ListenIp, config.ListenPort))
	if err != nil {
		log.Error("resolve tcp addr error", err)
		panic(err)
	}

	config.Listener, err = net.ListenTCP("tcp", listenAddr)
	if err != nil {
		log.Error("listen tcp addr error", err)
		panic(err)
	}

	log.Info("start server success")

	id := 0
	for {
		conn, err := config.Listener.AcceptTCP()
		if err != nil {
			if config.StopAccept {
				break
			}
			log.Error("accept tcp addr error", err)
			panic(err)
		}

		id++
		go newConn(id, conn)
	}

	log.Info("exit success")
}

func newConn(id int, conn *net.TCPConn) {
	clientAddr := conn.RemoteAddr().String()
	defer func() {
		_ = conn.Close()
		log.Debug(id, "close success:", clientAddr)
	}()
	log.Debug(id, "new connect:", clientAddr)

	clientConn := io.NewTCP(conn)
	clientConn.ServerInit(id)
	log.Info(id, "new client:", clientAddr)

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
	}(clientConn, readChan)

	// 心跳
	pingTicker := time.NewTicker(time.Duration(clientConn.Interval) * time.Second)
	defer pingTicker.Stop()
	lastPingTime := time.Now()
	lastPongTime := time.Now()

	var err error = nil
	for err == nil {
		select {
		case <-pingTicker.C:
			lastPingTime = time.Now()
			log.Trace(id, "send PingMessage")
			err = clientConn.WriteMessage(&msg.PingMessage{Date: lastPingTime})
			go func() {
				time.Sleep(10 * time.Second)
				if lastPongTime.Before(lastPingTime) {
					log.Warn(id, "ping timeout")
					_ = clientConn.Close() // 此处直接关闭连接, 让read线程退出方法
				}
			}()
		case message := <-readChan:
			if message.Err != nil {
				err = message.Err
				break
			}
			switch m := message.Message.(type) {
			case *msg.PoneMessage:
				log.Trace(id, "receiver Pong", m.Date)
				lastPongTime = time.Now()
			}
		}
	}

	log.Info(id, "exit:", err)
}
