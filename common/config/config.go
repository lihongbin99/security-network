package config

import (
	"flag"
	"gopkg.in/ini.v1"
	"security-network/common/logger"
)

var (
	log = logger.NewLog("Config")

	File    string
	IniFile *ini.File

	LogLevel string
)

func init() {
	flag.StringVar(&File, "c", File, "config file")
}

func Init() {
	cfg, err := ini.Load(File)
	if err != nil {
		panic(err)
	}

	IniFile = cfg

	// 获取公共配置
	LogLevel = IniFile.Section("log").Key("level").String()
	logger.SetLogLevel(LogLevel)
}
