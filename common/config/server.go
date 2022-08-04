package config

import (
	"io/ioutil"
	"net"
	"os"
	"strconv"
)

var (
	ListenIp   = ""
	ListenPort = 13520

	PrivateKey         = "config/security/private.pem"
	SecurityPrivateKey []byte

	TCPInterval = 180

	UDPInterval = 180

	Listener   *net.TCPListener
	StopAccept = false
)

func ServerInit() {
	if listenIp := IniFile.Section("listen").Key("ip").String(); listenIp != "" {
		ListenIp = listenIp
	}
	if listenPort := IniFile.Section("listen").Key("port").String(); listenPort != "" {
		port, err := strconv.Atoi(listenPort)
		if err != nil {
			log.Error("listen port error:", listenPort)
			panic(err)
		}
		ListenPort = port
	}

	if securityKey := IniFile.Section("security").Key("key").String(); securityKey != "" {
		_, err := os.Stat(securityKey)
		if err != nil {
			log.Error("not find security key:", securityKey)
			panic(err)
		}
		PrivateKey = securityKey
	}
	if temp, err := ioutil.ReadFile(PrivateKey); err != nil {
		log.Error("read private key pem error", err)
	} else {
		SecurityPrivateKey = temp
	}

	if tcpInterval := IniFile.Section("tcp").Key("timeout").String(); tcpInterval != "" {
		interval, err := strconv.Atoi(tcpInterval)
		if err != nil {
			log.Error("tcp interval error:", tcpInterval)
			panic(err)
		}
		TCPInterval = interval
	}

	if udpInterval := IniFile.Section("udp").Key("timeout").String(); udpInterval != "" {
		interval, err := strconv.Atoi(udpInterval)
		if err != nil {
			log.Error("tcp interval error:", udpInterval)
			panic(err)
		}
		UDPInterval = interval
	}

	log.Trace("listen ip   :", ListenIp)
	log.Trace("listen port :", ListenPort)
	log.Trace("security key:", PrivateKey)
	log.Trace("tcp timeout :", TCPInterval)
	log.Trace("udp timeout :", UDPInterval)
}

func StopServer() {
	StopAccept = true
	if Listener != nil {
		_ = Listener.Close()
	}
}
