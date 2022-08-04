package utils

import "fmt"

func I2b32(i uint32) (buf []byte) {
	buf = make([]byte, 4, 4)
	buf[0] = byte(i >> 24)
	buf[1] = byte(i >> 16)
	buf[2] = byte(i >> 8)
	buf[3] = byte(i)
	return
}

func B2i32(buf []byte) (i uint32, err error) {
	if len(buf) >= 4 {
		i = uint32(buf[0])<<24 +
			uint32(buf[1])<<16 +
			uint32(buf[2])<<8 +
			uint32(buf[3])
	} else {
		err = fmt.Errorf("id byte arr len = %d", len(buf))
	}
	return
}
