// Package pkg.
// creator 2022-03-30 18:36:16
// Author  zhenxinma.
package pkg

//import (
//	"context"
//	"fmt"
//	"io"
//	"log"
//
//	"gitee.com/zhenxinma/go_rpc/internal/pb"
//	"github.com/lucas-clemente/quic-go"
//)
//
//// Client Rpc 客户端.
//type Client struct {
//	Session quic.Session // 用于保持会话.
//	Stream  quic.Stream  // 用于实际传输数据.
//}
//
//func NewClient(addr string) (*Client, error) {
//
//	session, err := quic.DialAddr(addr, main.GetClientQuicTlsConfig(), nil)
//	if err != nil {
//		// 日志应该进行输出.
//		log.Println(err.Error())
//		return nil, err
//	}
//
//	stream, err := session.OpenStreamSync(context.Background())
//	if err != nil {
//		fmt.Println(err.Error())
//		return nil, err
//	}
//
//	data := []byte("Hello word")
//	stream.Write(data)
//
//	return &Client{
//		Session: session,
//	}, nil
//}
//
//// Open 打开QUIC传输通道Stream.
//func (c *Client) Open(con context.Context) (quic.Stream, error) {
//	// TODO: 最好返回一个Client.
//	return c.Session.OpenStreamSync(con)
//}
//
//// Call return the rpc bytes of result.
//// con 用来进行超时控制.
//func (c *Client) Call(req *pb.RpcRequest, con context.Context) ([]byte, error) {
//
//	//data, errcode := proto.Marshal(req)
//	//if errcode != nil {
//	//	return nil, errcode
//	//}
//
//	stream, err := c.Open(con)
//	if err != nil {
//		return nil, err
//	}
//
//	_, err = stream.Write([]byte("hello word"))
//	// 写完之后等待接受数据即可.
//	if err != nil {
//		return nil, err
//	}
//
//	// 等待读取数据.
//	// 先实现简单的功能,复杂的功能后续在进行慢慢改造.
//	accept := &main.Accept{}
//	_, err = io.Copy(accept, stream)
//	if err != nil {
//		return nil, err
//	}
//	// 返回本次远程调用所获取的所有数据.
//	return accept.Data, nil
//}
