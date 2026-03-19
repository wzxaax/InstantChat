package kernal

import (
	"encoding/binary"
	"io"
	"net"
)

/* 自定义消息结构体，按此规范进行网络I/O的读写，解决TCP消息边界问题 */

func ReadPack(conn net.Conn) ([]byte, error) {
	// 1. 读取数据包头部 4 字节的长度信息
	header := make([]byte, 4)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(header) // 解析长度 (大端序)

	// 2. 根据长度读取包体
	body := make([]byte, length)
	_, err = io.ReadFull(conn, body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func WritePack(conn net.Conn, data []byte) error {
	// 1. 准备长度前缀
	length := uint32(len(data))
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, length)

	// 2. 将前缀和包体合并发送
	packet := make([]byte, 4+len(data))
	copy(packet[0:4], header)
	copy(packet[4:], data)

	_, err := conn.Write(packet)
	return err
}
