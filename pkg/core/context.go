package core

import (
	"context"
	"net"
	"reflect"
	"sync"

	"github.com/xmopen/gorpc/pkg/errcode"
	"github.com/xmopen/gorpc/pkg/pack"
)

// RPCContext rpc Context Client/Server通用.
type RPCContext struct {
	context.Context
	mu   sync.RWMutex // rpc context 互斥锁.
	tags map[interface{}]interface{}
	conn net.Conn      // TCP CONN.
	Req  *pack.Message // req.
}

// NewRPCContext 构造RPC上下文.
func NewRPCContext(c context.Context, req *pack.Message) *RPCContext {
	return &RPCContext{
		Req:     req,
		Context: c,
	}
}

func (r *RPCContext) Lock() {
	r.mu.Lock()
}

func (r *RPCContext) Unlock() {
	r.mu.Unlock()
}

// SetConn set current conn.
func (r *RPCContext) SetConn(conn net.Conn) {
	r.conn = conn
}

// GetConn return current context conn.
func (r *RPCContext) GetConn() net.Conn {
	// if r.conn nil ?
	return r.conn
}

func (r *RPCContext) Value(key interface{}) interface{} {
	if r.tags == nil {
		r.tags = make(map[interface{}]interface{})
	}

	if val, ok := r.tags[key]; ok {
		return val
	}

	// 在尝试从context中获取.
	return r.Context.Value(key)
}

// SetValue set value for context.
func (r *RPCContext) SetValue(key, val interface{}) {
	r.Lock()
	defer r.Unlock()
	if r.tags == nil {
		r.tags = make(map[interface{}]interface{})
	}
	r.tags[key] = val
}

func (r *RPCContext) Delete(key interface{}) {
	r.Lock()
	defer r.Unlock()
	// 从map中删除掉
	if r.tags == nil || key == nil {
		return
	}
	delete(r.tags, key)
}

// WithValue init RPCContext.
func WithValue(parent context.Context, key, value interface{}) *RPCContext {
	if key == nil {
		panic("key is nil")
	}
	// Hash中的Key必须具有可比较性.
	if !reflect.TypeOf(key).Comparable() {
		panic("gorpc server rpcContext withValue kei is not comparable")
	}
	tags := make(map[interface{}]interface{})
	tags[key] = value
	return &RPCContext{tags: tags, Context: parent}
}

func WithRequest(c context.Context, req *pack.Message) *RPCContext {
	return &RPCContext{
		Context: c,
		Req:     req,
	}
}

// WriteResponse 响应请求.
func (r *RPCContext) WriteResponse(rsp *pack.Message) error {

	data, err := rsp.Encode()
	if err != nil {
		return err
	}

	index, err := r.conn.Write(*data)
	if err != nil {
		return err
	}
	if index != len(*data) {
		return errcode.SendResponseLengthNotFull
	}
	return nil
}

// TODO: WithTimeOut实现 Client超时.
