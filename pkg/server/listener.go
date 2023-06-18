package server

import (
	"fmt"
	"net"
)

var makeListener = make(map[string]MakeListener)

// MakeListener 通过指定的address构造net.Listener
type MakeListener = func(s *Server, address string) (net.Listener, error)

func init() {
	// 这个设计的好巧妙.
	makeListener["tcp"] = tcpMakeListener("tcp")
}

// tcpMakeListener 通过TCP构造连接.
func tcpMakeListener(network string) MakeListener {
	return func(s *Server, address string) (net.Listener, error) {
		listener, err := net.Listen(network, address)
		if err != nil {
			return nil, err
		}
		return listener, nil
	}
}

// createListener create net.Listener by net,address.
func (s *Server) createListener(network, address string) (net.Listener, error) {

	// network: tcp、udp.
	if ml, ok := makeListener[network]; ok {
		listener, err := ml(s, address)
		if err != nil {
			return nil, err
		}
		if listener == nil {
			return nil, fmt.Errorf("gorpc createListener listener is nil")
		}
		s.ln = listener
	}
	return s.ln, nil
}
