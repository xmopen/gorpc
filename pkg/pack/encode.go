package pack

import (
	"encoding/binary"
)

// 自定义协议
const (
	ProtocolMargic  = uint8(98) // 魔数.
	ProtocolVersion = uint8(8)
	ProtocolIsOne   = uint8(0) // 0、非单次请求 1、单次请求.

	ProtocolNullFiled = uint32(0) // 协议中为空的属性.
)

// RPC Request Type
const (
	RPCTypeConnType     RpcConnType = iota
	RPCTypePeerConnType             // 1
)

type RpcConnType = uint8

/*
1. 向字节数组中写入数据.
	// 存长度的时候就这么存.
	response := make([]byte, 0)
	binray.BigEndian.PutUint32(response,uint32(13))
*/

// Encode 编码.
// 返回字节指针,避免Copy内存.
func (m *Message) Encode() (*[]byte, error) {

	pl := len(m.Payload)
	dl := len(m.Data)
	sn := len(m.ServicePath)
	mn := len(m.ServiceMethod)
	total := defaultHeaderLength + pl + dl + sn + mn
	// 前20个字节是编码,将当前Message相关内容编码到
	response := make([]byte, 0, total)
	// 0-4个字节基础信息.
	response = append(response, ProtocolMargic)
	response = append(response, ProtocolVersion)
	response = append(response, m.Type)
	// 第4个字节控制Timeout.
	response = append(response, uint8(m.Timeout.Seconds()))
	//response = append(response, ProtocolIsOne)
	// 4-8个字节 payloadLen
	array := make([]byte, 4)
	binary.BigEndian.PutUint32(array, uint32(len(m.Payload)))
	response = append(response, array...)
	// 8-12个字节 dataLen
	array = make([]byte, 4)
	binary.BigEndian.PutUint32(array, uint32(len(m.Data)))
	response = append(response, array...)

	// 12-16个字节 serverNameLen
	array = make([]byte, 4)
	binary.BigEndian.PutUint32(array, uint32(len([]byte(m.ServicePath))))
	response = append(response, array...)

	// 16-20个字节 methodNameLen
	array = make([]byte, 4)
	binary.BigEndian.PutUint32(array, uint32(len([]byte(m.ServiceMethod))))
	response = append(response, array...)

	array = make([]byte, 0)
	array = append(array, []byte(m.ServicePath)...)
	array = append(array, []byte(m.ServiceMethod)...)
	array = append(array, m.Payload...)
	array = append(array, m.Data...)
	response = append(response, array...)
	res := response[:total]
	return &res, nil
}
