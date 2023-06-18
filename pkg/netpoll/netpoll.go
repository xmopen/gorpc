package netpoll

import (
	"net"
)

const (
	TCP = "tcp"
)

// 先独立出来,后续有可能会会自己独立封装网络库.

type NetFuncCall = func(network string) (net.Conn, error)

var netCall = map[string]NetFuncCall{
	TCP: netTcp,
}

// 默认支持TCP.
func netTcp(addr string) (net.Conn, error) {
	ln, err := net.Dial(TCP, addr)
	if err != nil {
		return nil, err
	}
	return ln, err
}

// Dial return net.conn.
func Dial(network string) NetFuncCall {
	return netCall[network]
}
