package pack

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/xmopen/gorpc/pkg/util"
)

const (
	defaultHeaderLength = 20
)

// 解码.

// Decode Message Decode.
// []byte ---> Message Struct.
// TODO: 待优化:使用对象池.
func (m *Message) Decode(req io.Reader) error {

	// 这里提前读取Header是应为如果本次请求是不安全的是伪造的,那么后面就不需要去解析了.
	//header := make([]byte, defaultHeaderLength)
	n, err := req.Read(m.Header.Source[:20])
	if err != nil {
		// client主动关闭 server端则会读取一个EOR.
		return err
	}
	if n < defaultHeaderLength {
		return fmt.Errorf("req length < defaultHeaderLength(12)")
	}

	// 后续可以通过性能测试的时候进行改善.

	// 读取长度不应该从Header中读取吗?
	// 0-4 个字节是魔数等(暂时不进行考虑.).
	// 4-8个字节是payLoadLen 就是请求参数长度.
	// 8-12个字节是dataLen本次响应长度.
	// 12-16个字节是servicename的长度.
	// 16-20个字节是methodname的长度.
	// 不会出现重复读取现象.

	// pll->payloadLength
	pll := binary.BigEndian.Uint32(m.Header.Source[4:8])
	// dl->dataLength
	dl := binary.BigEndian.Uint32(m.Header.Source[8:12])
	// snl->serviceNameLength
	snl := binary.BigEndian.Uint32(m.Header.Source[12:16])
	// ml->methodLength
	ml := binary.BigEndian.Uint32(m.Header.Source[16:20])
	//start := defaultHeaderLength
	m.Timeout = time.Duration(int8(m.Header.Source[3])) * time.Second
	m.Type = m.Header.Source[2]

	// 1. 读取服务名称内容.
	servicePathBytes := make([]byte, snl)
	n, err = req.Read(servicePathBytes)
	//_, err = io.ReadFull(req, servicePathBytes)
	if err != nil {
		return err
	}
	if n != int(snl) {
		return fmt.Errorf("gorpc server read serverPath length err")
	}
	m.ServicePath = util.BytesSliceToString(servicePathBytes)

	// 2. 读取本次调用方法内容.
	methodNameBytes := make([]byte, ml)
	n, err = req.Read(methodNameBytes)
	if err != nil {
		return err
	}
	if n != int(ml) {
		return fmt.Errorf("gorpc server read server method info length err")
	}
	m.ServiceMethod = util.BytesSliceToString(methodNameBytes)

	// PayLoad字节数组
	m.Payload = make([]byte, pll)
	n, err = req.Read(m.Payload)
	if err != nil {
		return err
	}
	if n != int(pll) {
		return fmt.Errorf("gorpc server read payload length err")
	}

	m.Data = make([]byte, dl)
	n, err = req.Read(m.Data)
	if err != nil {
		return err
	}
	if n != int(dl) {
		return fmt.Errorf("gorpc server read data length err")
	}
	return nil
}

// metaBytes2Map 将字节数组表示的MetaData转换成Map.
func metaBytes2Map(data []byte) (map[string]string, error) {
	metaMap := make(map[string]string)
	buf := bytes.NewBuffer(data)
	if err := binary.Read(buf, binary.LittleEndian, metaMap); err != nil {
		return metaMap, err
	}
	return metaMap, nil
}
