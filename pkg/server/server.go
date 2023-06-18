// Package server rpc server.
package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"runtime"
	"sync"
	"time"

	"gitee.com/zhenxinma/gocommon/pkg/xlogging"
	"gitee.com/zhenxinma/gorpc/pkg/core"
	"gitee.com/zhenxinma/gorpc/pkg/pack"
)

const (
	RpcContextRemoteAddrKey = "gorpc remote addr"

	BufferReaderSize = 1024 // 1K.
)

var (
	typeOfError   = reflect.TypeOf((*error)(nil)).Elem()
	typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()
)

var PublicServer *Server

type OptionFunc func(server *Server)

// Server Server.
// 属性不主动对外进行暴露,通过OptionFunc以及Set来进行设置.
type Server struct {
	// net.listener 配置.
	ln           net.Listener // tcp listener.
	readTimeout  time.Duration
	writeTimeout time.Duration
	options      map[string]interface{}

	// service 锁.
	serviceMapMu sync.RWMutex
	serviceMap   map[string]*service // 每一个对外进行暴露的服务都必须注册在该map中.

	router map[string]interface{} // todo.

	// Server锁.
	mu sync.RWMutex
	// 下面的不是很理解.
	activeConn map[net.Conn]struct{}
	doneChan   chan struct{}
	seq        uint64

	trace bool // 默认false.
	xlog  *xlogging.Entry

	Plugins PluginContainer
}

// NewServer 初始化Server.
func NewServer(options ...OptionFunc) *Server {

	s := &Server{
		Plugins:    &pluginContainer{},
		options:    make(map[string]interface{}),
		activeConn: make(map[net.Conn]struct{}),
		doneChan:   make(chan struct{}),
		serviceMap: make(map[string]*service),
		router:     make(map[string]interface{}),
		xlog:       xlogging.Tag("gorpc server"),
		trace:      true,
	}

	for _, option := range options {
		// 对s的初始化进行干预.
		// 比如向s.options设置初始值.
		option(s)
	}

	// 先使用TCP. 最后能够做成可替换的.
	// 后期使用UDP.
	// 不管用啥其实最终我们用到的都是ln.
	if s.options["TCPKeepAlivePeriod"] == nil {
		s.options["TCPKeepAlivePeriod"] = 3 * time.Minute
	}

	PublicServer = s
	return s
}

// Server starts and listens RPC requests.
// It is blocked until receiving connections from clients
func (s *Server) Server(network, addr string) error {
	if err := s.Plugins.CustomPlugDo(s); err != nil {
		return err
	}
	ln, err := s.createListener(network, addr)
	if err != nil {
		return err
	}
	// serverListener启动链接.
	return s.serverListener(ln)
}

// serverListener accepts incoming connections on the Listener ln,
// creating a new service goroutine for each.
// The service goroutines read requests and then call services to reply to them.
func (s *Server) serverListener(ln net.Listener) error {

	s.mu.Lock()
	s.ln = ln
	s.mu.Unlock()

	var tempDelay time.Duration
	// start tcp conn.
	for {
		conn, err := s.ln.Accept()
		// err 解决.
		if err != nil {
			// 超时问题.
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					// 第一次等待5毫秒.
					tempDelay = 5 * time.Millisecond
				} else {
					// 二倍等待时间.
					tempDelay *= 2
				}
				// 对最大等待时间进行限制.
				if tempDelay > 1*time.Second {
					tempDelay = 1 * time.Second
				}
				time.Sleep(tempDelay)
				continue
			}
			// todo: err.
		}

		// 正常应该如何进行处理.
		if tc, ok := conn.(*net.TCPConn); ok {
			// 设置心跳保活.
			period := s.options["TCPKeepAlivePeriod"]
			if period != nil {
				err = tc.SetKeepAlive(true)
				if err != nil {

				}

				tc.SetKeepAlivePeriod(period.(time.Duration))
				tc.SetLinger(10)
			}
		}

		if s.trace {
			// conn.RemoteAddr().String() return 127.0.0.1.
			s.xlog.Debugf("go.rpc server accept conn:%v", conn.RemoteAddr().String())
		}

		// 开启一个goroutine进行处理.
		go s.server(conn)
	}
}

// server 核心处理逻辑.
func (s *Server) server(conn net.Conn) {

	defer func() {
		if err := recover(); err != nil {
			// panic.
			// 2的16次方.
			size := 64 << 10
			buf := make([]byte, size)
			ss := runtime.Stack(buf, false)
			if ss > size {
				ss = size
			}
			buf = buf[:ss]
			s.xlog.Errorf("server conn:%v is panic err:%v msg:%v", conn.RemoteAddr().String(), err, string(buf))
		}

		if s.trace {
			s.xlog.Debugf("server closed conn:%v", conn.RemoteAddr().String())
		}

		s.closeConn(conn)
	}()

	now := time.Now()
	if s.readTimeout != 0 {
		conn.SetReadDeadline(now.Add(s.readTimeout))
	}

	ctx := core.WithValue(context.Background(), RpcContextRemoteAddrKey, conn.RemoteAddr().String())
	ctx.SetConn(conn)

	// 先不设置大小了吧.
	buffer := bufio.NewReaderSize(conn, BufferReaderSize)
	for {

		// 粘包,粘包解决.
		// read是阻塞的,所以这里在读取了一个request之后就go response.
		// 判断当前消息是否读取完,是根据长度来记性读取的.
		// 假如A在很短的时间内发送了两个request,那么通过阻塞和for循环依旧能够读取出每一个request.
		req, err := s.readMessage(ctx, buffer, conn)
		if err != nil {
			if err == io.EOF {
				// 正常结束.再次进入循环中,如果上一次已经读取完数据,这里将会出现EOF.
				if s.trace {
					s.xlog.Debugf("client closed the conn:%v", conn.RemoteAddr().String())
				}
			} else if errors.Is(err, net.ErrClosed) {
				// 向已经关闭的conn中写入数据则会发生ErrClosed.
				s.xlog.Errorf("conn:%v is closed", conn.RemoteAddr().String())
			}
			return
		}

		ctx.Req = req
		go s.response(ctx)
	}
}

// readMessage read data from tcp to pack.Message.
func (s *Server) readMessage(ctx *core.RPCContext, buffer *bufio.Reader, conn net.Conn) (*pack.Message, error) {

	req := pack.NewPackMessage()
	err := req.Decode(conn)
	if err != nil {
		//s.xlog.Errorf("req.Decode err:%v", err)
		return nil, err
	}
	return req, nil
}

func (s *Server) closeConn(con net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	con.Close()
}

// SetTrace 设置Server是否输出详细日志默认为False.
// 其实应该还有日志级别设置的.
func (s *Server) SetTrace(trace bool) {
	if trace {
		s.xlog = xlogging.Tag(" go_rpc_server")
	}
}
