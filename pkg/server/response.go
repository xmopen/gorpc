package server

import (
	"encoding/json"
	"fmt"
	"reflect"
	"runtime/debug"

	"github.com/xmopen/golib/pkg/xlogging"

	"github.com/xmopen/gorpc/pkg/core"
	"github.com/xmopen/gorpc/pkg/errcode"
	"github.com/xmopen/gorpc/pkg/pack"
)

var (
	ServerNotFound  = fmt.Errorf("ServerNotFound")
	MethodNotFound  = fmt.Errorf("MethodNotFound")
	ResponseDataNil = fmt.Errorf("ResponseDataIsNil")
)

// response 处理响应.
func (s *Server) response(ctx *core.RPCContext) {
	defer func() {
		if err := recover(); err != nil {
			xlogging.Tag("gorpc.response").Errorf("panic:[%+v]", string(debug.Stack()))
		}
	}()

	if s.trace {
		s.xlog.Debugf("gorpc server response req:%+v", ctx.Req)
	}

	// 正常处理.
	rsp, err := s.handleRequest(ctx)
	if err != nil {
		if s.trace {
			s.xlog.Errorf("handleRequest err:%v", err)
		}
		rsp = pack.NewPackMessage()
		rsp.Data = []byte(err.Error())
		rsp.Type = uint8(pack.RPCRequestTypeOfResponseError)
	}
	err = ctx.WriteResponse(rsp)
	if err != nil {
		s.xlog.Errorf("service: %s method:%s write response err:%v", ctx.Req.ServicePath, ctx.Req.ServiceMethod, err)
	}
}

// handleRequest 处理请求.
func (s *Server) handleRequest(ctx *core.RPCContext) (*pack.Message, error) {

	serviceName := ctx.Req.ServicePath
	methodName := ctx.Req.ServiceMethod

	server, ok := s.serviceMap[serviceName]
	if !ok {
		return nil, ServerNotFound
	}
	method, ok := server.method[methodName]
	if !ok {
		return nil, MethodNotFound
	}

	// TODO: 目前请求参数和响应参数都必须是指针.
	argv := reflect.New(method.ArgType.Elem()).Interface()
	e := json.Unmarshal(ctx.Req.Payload, &argv)
	if e != nil {
		return nil, e
	}

	var rspv reflect.Value
	if method.replyType.Kind() == reflect.Pointer {
		rspv = reflect.New(method.replyType.Elem())
	} else {
		return nil, errcode.ServerMethodResponseNotPointer
	}

	// response都已经是指针了啊,字段还有必要是指针？
	rspvItr := rspv.Interface()
	var err error
	// TODO: 参数可以是指针也可以不是指针,但是返回值必须是指针.
	if method.ArgType.Kind() != reflect.Pointer {
		err = server.Call(ctx, method, reflect.ValueOf(argv), reflect.ValueOf(rspvItr))
	} else {
		err = server.Call(ctx, method, reflect.ValueOf(argv), reflect.ValueOf(rspvItr))
	}
	if err != nil {
		s.xlog.Errorf("method call err:%v", err)
		return nil, err
	}
	resp := pack.NewPackMessage()
	data, err := json.Marshal(rspvItr)
	if err != nil {
		return nil, err
	}
	resp.Data = data
	resp.Type = pack.MessageTypePeerRPCResponse
	return resp, nil
}
