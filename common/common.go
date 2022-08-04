package common

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	AppName    string
	AppVersion string
	Protocol   []byte
)

func init() {
	AppName = "security-network"
	AppVersion = "1.0.0"

	split := strings.Split(AppVersion, ".")
	if len(split) != 3 {
		panic(fmt.Errorf("AppVersion error: %s", AppVersion))
	}

	version := make([]byte, 3)
	for i := 0; i < 3; i++ {
		v, err := strconv.Atoi(split[i])
		if err != nil {
			panic(fmt.Errorf("AppVersion error: %s", AppVersion))
		}
		version[i] = byte(v)
	}

	Protocol = make([]byte, 0)
	Protocol = append(Protocol, AppName...)
	Protocol = append(Protocol, version...)
}
