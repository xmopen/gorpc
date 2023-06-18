package server

import (
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/xmopen/gorpc/pkg/core"
)

// Service 服务具体结构.
// 一个Service表示一个服务.
type service struct {
	name     string        // 服务名字.
	serType  reflect.Type  // Server Type.
	serValue reflect.Value // Server Value.
	method   map[string]*methodType
	function map[string]*functionType
}

// methodType method抽象.
type methodType struct {
	sync.Mutex
	method    reflect.Method
	ArgType   reflect.Type
	replyType reflect.Type
}

// functionType function抽象.
type functionType struct {
	sync.Mutex
	fn        reflect.Value
	ArgType   reflect.Type
	replyType reflect.Type
}

// Call 调用本地服务方法.
func (s *service) Call(ctx *core.RPCContext, methodType *methodType, arg, rsp reflect.Value) error {

	function := methodType.method.Func
	// 第一个参数表示具体的Service的Value.
	// 第二个参数表示上下文.
	// 第三个参数和第四个参数分别表示该方法的入参.
	// 其实可以由多个返回值,但是不确定返回几个,所有默认通过Error来进行返回.
	// 其实执行Call就已经将rsp赋值了.
	returnValues := function.Call([]reflect.Value{s.serValue, reflect.ValueOf(ctx.Context), arg, rsp})
	err := returnValues[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}

// isExported 判断name是否为可导的类型,就是判断name第一个字符是否为大写.
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}
