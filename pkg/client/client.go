package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"time"

	"github.com/xmopen/golib/pkg/xlogging"
	"github.com/xmopen/gorpc/pkg/errcode"
	"github.com/xmopen/gorpc/pkg/netpoll"
	"github.com/xmopen/gorpc/pkg/pack"
)

var DefaultTimeOut = 15 * time.Second
var DefaultServerTimeout = 10 * time.Second

// Client RPC Client.
type Client struct {
	Trace         bool          // 是否打印Client日志.
	ServerTimeout time.Duration // 服务端超时控制.
	ClientTimeout time.Duration // 客户端超时控制.
	Xlog          *xlogging.Entry
	PoolConfig    *netpoll.PoolConfig // pool config.

	network string   // 网络类型.
	addr    string   // Server addr.
	conn    net.Conn // response.

	cancel context.CancelFunc
	err    chan error
	signal chan struct{} // signal 用于进行信号传输,空结构体占用0字节.

}

// OptionConfig client config.
type OptionConfig struct {
	ServerTimeout time.Duration // 服务端超时控制.
	ClientTimeout time.Duration // 客户端超时时间.
	Xlog          *xlogging.Entry
}

// NewClient 返回一个客户端,如果err为空,则client则是可用的.
// TODO: 连接池.
func NewClient(network, addr string, cfg *OptionConfig) (*Client, error) {

	c := &Client{
		ClientTimeout: DefaultTimeOut,
		ServerTimeout: DefaultServerTimeout,
		network:       network,
		addr:          addr,
		err:           make(chan error, 1), // 容量为1不会阻塞.
		signal:        make(chan struct{}, 1),
		Xlog:          xlogging.Tag("gorpc.client"),
	}

	if cfg != nil {
		if cfg.ClientTimeout != 0 {
			c.ClientTimeout = cfg.ClientTimeout
		}
		if cfg.ServerTimeout != 0 {
			c.ServerTimeout = cfg.ServerTimeout
		}

		if cfg.Xlog != nil {
			c.Xlog = cfg.Xlog
		}
	}

	return c, nil
}

func (c *Client) setConn(conn net.Conn) {
	c.conn = conn
}

// Close 关闭客户端.
func (c *Client) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.Xlog.Errorf("gorpc client close err:%v", err)
		}
	}
}

// Call 发起远程调用.
// Call 不应该是最上层,或者说当前Call实现内容不应该是最上层.
func (c *Client) Call(ctx context.Context, serviceName string, methodName string, args interface{}, resp interface{}) error {

	defer c.Close()
	start := time.Now().Unix()
	defer func(st int64) {
		end := time.Now().Unix()
		if c.Trace {
			c.Xlog.Infof("servicename:%v methodname:%v calltime:%ds result:%+v", serviceName, methodName,
				end-start, resp)
		}
	}(start)
	ctx, cancel := context.WithTimeout(ctx, c.ClientTimeout)
	c.cancel = cancel
	go c.Go(ctx, serviceName, methodName, args, resp)
	// 真实的做法是将timeout透传给server端,server端在处理时间超过给定值之后必须进行write,且最后write时必须检查当前是否已经write过.
	select {
	case <-ctx.Done():
		return errcode.ClientCallTimeOut
	case e := <-c.err:
		return e
	case <-c.signal:
		// 结束本次Call().
		return nil
	}
}

// Go 具体远程调用实现逻辑.
func (c *Client) Go(ctx context.Context, sn string, mn string, args interface{}, resp interface{}) {

	defer func() {
		if err := recover(); err != nil {
			//发生panic.
			c.Xlog.Errorf("client.Go recover err:%v", err)
			c.DoneError(errcode.ClientCallPanicError)
		}
	}()

	err := c.send(ctx, sn, mn, args, resp)
	// 如何将Timeout转换成int.
	if err != nil {
		c.Xlog.Errorf("gorpc client.send err:%v", err)
		c.DoneError(err)
		return
	}

	// 读取本次RPC请求响应.
	// 服务端并不会主动向RPC端推送消息.
	// 所以这里目前只会接受一个消息.
	err = c.readResponse(resp)
	if err != nil {
		c.Xlog.Errorf("gorpc client read response err:%v", err)
		c.DoneError(err)
		return
	}
	// 正常请求完成,发送信号量通知上层本次请求结束.
	c.Done()
}

// send 发送数据.
func (c *Client) send(ctx context.Context, sn, mn string, args, resp interface{}) error {

	req := pack.NewPackMessage()
	ab, err := json.Marshal(args)
	if err != nil {
		return err
	}
	req.Payload = ab

	rb, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	req.Data = rb
	req.ServicePath = sn
	req.ServiceMethod = mn
	req.Timeout = c.ServerTimeout // 服务端超时控制.
	req.Type = pack.MessageTypePeerRPCRequest

	data, err := req.Encode()
	if err != nil {
		return err
	}
	// todo: 连接池.
	ln, err := netpoll.Dial(c.network)(c.addr)
	c.Xlog.Debugf("client dial addr:%v", c.addr)
	if err != nil {
		return err
	}

	n, err := ln.Write(*data)
	if err != nil {
		return err
	}
	if n != len(*data) {
		return err
	}
	c.setConn(ln)
	return nil
}

// readResponse client read response from server.
func (c *Client) readResponse(response interface{}) error {

	// 客户端主动关闭,Server端读取数据会发生EOF.
	resp := pack.NewPackMessage()
	err := resp.Decode(c.conn)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// server端主动关闭.
			if c.Trace {
				c.Xlog.Errorf("the gorpc server initiative tcp conn")
			}
		} else if errors.Is(err, net.ErrClosed) {
			// 1、读取已经关闭的TCP连接.
			// 2、当前goroutine还没有读取TCP连接中的数据,其他goroutine就已经关闭TCP连接了.
			c.Xlog.Errorf("conn is not available, it may have been closed or closed by another goroutine err:%v", err)
		}
		return err
	}

	// 其实这里应该责任链控制.
	// 是否超时控制.
	if resp.Type == pack.MessageTypeServerTimeout {
		return errcode.ClientCallTimeOut
	}

	if len(resp.Data) > 0 {
		err = json.Unmarshal(resp.Data, response)
		if err != nil {
			return err
		}
	}
	c.Xlog.Debugf("client readResponse result:%+v", response)
	return nil
}

// Done Client通信.
func (c *Client) Done() {
	c.signal <- struct{}{}
}

// DoneError 通过发送err结束本次client.Call流程调用.
func (c *Client) DoneError(err error) {
	c.err <- err
}
