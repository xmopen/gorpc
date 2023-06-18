package server

import (
	"context"
	"sync"
)

type Context struct {
	lock *sync.Mutex
	tags map[interface{}]interface{}
	context.Context
}

// NewContext return a new Context of me.
func NewContext(ctx context.Context) *Context {
	tagLock := &sync.Mutex{}
	ctx = context.WithValue(ctx, "", tagLock)
	return &Context{
		lock:    tagLock,
		Context: ctx,
		tags:    make(map[interface{}]interface{}),
	}
}
