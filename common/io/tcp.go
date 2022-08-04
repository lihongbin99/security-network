package io

import (
	"fmt"
	"net"
	"security-network/common"
	"security-network/common/config"
	"security-network/common/msg"
	"security-network/common/security"
	"security-network/common/utils"
	"strconv"
	"time"
)

type TCP struct {
	*net.TCPConn
	Buf      []byte
	Key      []byte
	Iv       []byte
	Interval int
}

func NewTCP(conn *net.TCPConn) *TCP {
	t := &TCP{conn, make([]byte, 64*1024), nil, nil, 0}
	return t
}

func (that *TCP) ReadByLen(maxReadLen int, timeout time.Time) ([]byte, error) {
	_ = that.SetReadDeadline(timeout)

	buf := make([]byte, maxReadLen)
	readSum := 0

	for readSum < maxReadLen {
		readLen, err := that.Read(buf[readSum:])
		if err != nil {
			return nil, err
		}
		readSum += readLen
	}

	_ = that.SetReadDeadline(time.Time{})
	return buf, nil
}

func (that *TCP) ServerInit(id int) (success bool) {
	// 接受协议
	protocol, err := that.ReadByLen(len(common.Protocol), time.Now().Add(8*time.Second))
	if err != nil {
		log.Debug(id, "read protocol error", err)
		return
	}
	protocolLen := len(protocol)
	if string(protocol[:protocolLen-3]) != common.AppName {
		log.Debug(id, "protocol error")
		return
	}
	clientVersion :=
		strconv.Itoa(int(protocol[protocolLen-3])) + "." +
			strconv.Itoa(int(protocol[protocolLen-2])) + "." +
			strconv.Itoa(int(protocol[protocolLen-1]))
	log.Trace(id, "client version:", clientVersion)

	// 返回协议结果
	if _, err = that.Write(common.Protocol); err != nil {
		log.Debug(id, "write protocol error", err)
		return
	}

	// 接受密钥
	message, err := that.ReadByLen(256, time.Now().Add(8*time.Second))
	if err != nil {
		log.Debug(id, "read key error", err)
		return
	}
	message, err = security.DecryptRSA(message, config.SecurityPrivateKey)
	if err != nil {
		log.Debug(id, "DecryptRSA error", err)
		return
	}
	if len(message) != 32 {
		log.Debug(id, "key error len:", len(message))
		return
	}
	that.Key = message[:16]
	that.Iv = message[16:]

	// 返回密钥结果(心跳时间)
	that.Interval = config.TCPInterval
	encrypt, err := security.AesEncrypt([]byte(strconv.Itoa(that.Interval)), that.Key, that.Iv)
	if err != nil {
		log.Debug(id, "aes encrypt error", err)
		return
	}
	if _, err = that.Write(encrypt); err != nil {
		log.Debug(id, "write interval error", err)
		return
	}

	return true
}

func (that *TCP) ClientInit() (success bool) {
	// 发送协议
	if _, err := that.Write(common.Protocol); err != nil {
		log.Error("write protocol error", err)
		return
	}

	// 接受协议结果
	protocol, err := that.ReadByLen(len(common.Protocol), time.Now().Add(8*time.Second))
	if err != nil {
		log.Debug("read protocol error", err)
		return
	}
	protocolLen := len(protocol)
	if string(protocol[:protocolLen-3]) != common.AppName {
		log.Debug("protocol error")
		return
	}
	serverVersion :=
		strconv.Itoa(int(protocol[protocolLen-3])) + "." +
			strconv.Itoa(int(protocol[protocolLen-2])) + "." +
			strconv.Itoa(int(protocol[protocolLen-1]))
	log.Trace("server version:", serverVersion)

	// 发送密钥
	that.Key, that.Iv = security.GenerateAES()
	message := make([]byte, 32)
	copy(message[0:16], that.Key)
	copy(message[16:], that.Iv)
	message, err = security.EncryptRSA(message, config.SecurityPublicKey)
	if err != nil {
		log.Error("EncryptRSA error", err)
		return
	}
	if _, err = that.Write(message); err != nil {
		log.Error("write key error", err)
		return
	}

	// 接收密钥结果(心跳时间)
	message, err = that.ReadByLen(16, time.Now().Add(8*time.Second))
	if err != nil {
		log.Debug("read key result error", err)
		return
	}
	decrypt, err := security.AesDecrypt(message, that.Key, that.Iv)
	if err != nil {
		log.Debug("aes decrypt error", err)
		return
	}
	interval, err := strconv.Atoi(string(decrypt))
	if err != nil {
		log.Debug("interval error", err)
		return
	}
	that.Interval = interval

	return true
}

func (that *TCP) WriteMessage(message msg.Message) error {
	// 解析
	data, err := msg.ToByte(message)
	if err != nil {
		return err
	}
	if len(data) <= 0 {
		return nil
	}

	// 加密
	if data, err = security.AesEncrypt(data, that.Key, that.Iv); err != nil {
		return err
	}

	// 发送消息类型
	if _, err = that.Write(utils.I2b32(message.GetMessageType())); err != nil {
		return err
	}
	// 发送消息长度
	if _, err = that.Write(utils.I2b32(uint32(len(data)))); err != nil {
		return err
	}
	// 发送消息
	if _, err = that.Write(data); err != nil {
		return err
	}
	return nil
}

func (that *TCP) ReadMessage(timeout time.Time) Message {
	_ = that.SetReadDeadline(timeout)

	// 读取前缀
	readSum := 0
	for readSum < 8 {
		if readLength, err := that.Read(that.Buf[readSum:8]); err != nil {
			return Message{Err: err}
		} else {
			readSum += readLength
		}
	}

	// 获取消息类型
	messageType, err := utils.B2i32(that.Buf[:4])
	if err != nil {
		return Message{Err: err}
	}
	message, err := msg.NewMessage(messageType)
	if err != nil {
		return Message{Err: err}
	}

	// 获取消息长度
	messageLen32, err := utils.B2i32(that.Buf[4:8])
	if err != nil {
		return Message{Err: err}
	}
	messageLen := int(messageLen32)
	if messageLen <= 0 {
		return Message{Err: err}
	}
	if messageLen > len(that.Buf) {
		return Message{Err: fmt.Errorf("message len: %d", messageLen)}
	}

	// 读取消息
	readSum = 0
	for readSum < messageLen {
		if readLength, err := that.Read(that.Buf[readSum : messageLen-readSum]); err != nil {
			return Message{Err: err}
		} else {
			readSum += readLength
		}
	}
	data := that.Buf[:messageLen]

	// 解密
	if data, err = security.AesDecrypt(data, that.Key, that.Iv); err != nil {
		return Message{Err: err}
	}

	// 解析
	if err = msg.ToObj(data, message); err != nil {
		return Message{Err: err}
	}

	_ = that.SetReadDeadline(time.Time{})
	return Message{Message: message, Err: nil}
}
