package logger

import (
	"fmt"
	"strings"
	"time"
)

var (
	logLevelMap = map[string]int{
		"off":   0,
		"error": 1,
		"warn":  2,
		"info":  3,
		"debug": 4,
		"trace": 5,
	}
	level = logLevelMap["info"]

	logger = NewLog("Log")
)

func SetLogLevel(logLevel string) {
	if logLevel != "" {
		logLevel = strings.ToLower(logLevel)

		if l, exist := logLevelMap[logLevel]; exist {
			level = l
		} else {
			i := 0
			keys := make([]string, len(logLevelMap))
			for key := range logLevelMap {
				keys[i] = key
				i++
			}
			panic(fmt.Errorf("log level error: \"%s\" not in %v", logLevel, keys))
		}

		logger.Debug("log level:", logLevel)
	}
}

func getTime() string {
	return time.Now().Format("2006:01:02 15:04:05")
}
