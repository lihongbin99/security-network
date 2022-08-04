package config

import (
	"io/ioutil"
	"os"
	"strconv"
)

var (
	ServerIp   = "0.0.0.0"
	ServerPort = 13520

	ClientName = "default"

	PublicKey         = "config/security/public.pem"
	SecurityPublicKey []byte
)

func ClientInit() {
	if serverIp := IniFile.Section("server").Key("ip").String(); serverIp != "" {
		ServerIp = serverIp
	}
	if serverPort := IniFile.Section("server").Key("port").String(); serverPort != "" {
		port, err := strconv.Atoi(serverPort)
		if err != nil {
			log.Error("server port error:", serverPort)
			panic(err)
		}
		ServerPort = port
	}

	if clientName := IniFile.Section("client").Key("name").String(); clientName != "" {
		ClientName = clientName
	}

	if securityKey := IniFile.Section("security").Key("key").String(); securityKey != "" {
		_, err := os.Stat(securityKey)
		if err != nil {
			log.Error("not find security key:", securityKey)
			panic(err)
		}
		PublicKey = securityKey
	}
	if temp, err := ioutil.ReadFile(PublicKey); err != nil {
		log.Error("read public key pem error", err)
	} else {
		SecurityPublicKey = temp
	}

	log.Trace("server ip   :", ServerIp)
	log.Trace("server port :", ServerPort)
	log.Trace("security key:", PublicKey)
}
