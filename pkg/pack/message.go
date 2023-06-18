package pack

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	MessageTypeServerTimeout   = uint8(127) // 服务端超时响应,客户端不去控制超时.
	MessageTypePeerRPCRequest  = uint8(126) // 请求类型为单次RPC请求响应.
	MessageTypePeerRPCResponse = uint8(125)
)

// HeaderSource 消息头源.
type HeaderSource [20]byte

// Message RPC Message.
type Message struct {
	*Header // Header12个字节长度的固定格式.
	*MetaData
	MetaDataSource []byte // Method.
	Payload        []byte // Args.
	Data           []byte // Rsp.
}

type Header struct {
	Source     [20]byte
	Margic     uint8 // 魔数.
	Version    uint8 // version.
	Type       uint8 // 消息类型.
	IsOne      uint8 // 单次RPC.
	PayLoadLen uint8
	DataLen    uint8
}

// MetaData 源信息.
type MetaData struct {
	ServicePath   string        `json:"servicePath"`   // 具体那个服务.
	ServiceMethod string        `json:"serviceMethod"` // 具体那个服务下的那个方法.
	Timeout       time.Duration `json:"timeout"`       // 超时控制时间.
}

// NewPackMessage 构造一个新的Message.
func NewPackMessage() *Message {
	return &Message{
		MetaData: &MetaData{
			Timeout: 10 * time.Second,
		},
		Header: &Header{
			Source: [defaultHeaderLength]byte{},
		},
	}
}

// CheckMagic check req magic.
func (m *Message) CheckMagic() bool {
	xlog.Debugf("tcp message magic:%b magin:%v", m.Header.Source[0], m.Header.Margic)
	// TODO: 字节数组中保存的是16进制?
	if m.Header.Margic == MagicNumber {
		return true
	}
	return false
}

// CheckVersion check tcp req version.
func (m *Message) CheckVersion() bool {
	xlog.Debugf("tcp message version:%b version:%v", m.Header.Source[1], m.Header.Version)
	// TODO: 这里进行判断也应该是有问题的.
	if m.Header.Version == ProtocolVersion {
		return true
	}
	return false
}

// GetPayloadLength return tcp payload len.
func (m *Message) GetPayloadLength() int32 {
	return int32(binary.BigEndian.Uint32(m.Header.Source[4:8]))
}

func (m *Message) GetDataLength() int32 {
	return int32(binary.BigEndian.Uint32(m.Header.Source[8:12]))
}

func (m *Message) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ByteToStruct(data []byte) (*Message, error) {
	buf := new(bytes.Buffer)
	m := NewPackMessage()
	if err := binary.Read(buf, binary.BigEndian, m); err != nil {
		return nil, err
	}
	return m, nil
}
