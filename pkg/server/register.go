package server

import (
	"fmt"
	"reflect"
)

const (
	serverMethodArgsIndex        int = iota
	serverMethodContentArgsIndex     // RPC上下文.
	serverMethodReqArgsIndex         // 本次请求参数.
	serverMethodRspArgsIndex         // 本次相应参数.
	rpcMethodTotalArgs
)

// RegisterName 对外公开注册函数,metadata用于其他插件使用.
func (s *Server) RegisterName(name string, server interface{}, metadata string) error {

	// 这里不应该new的.
	_, err := s.register(name, server, true)
	if err != nil {
		return err
	}
	if s.Plugins == nil {
		s.Plugins = &pluginContainer{}
	}

	// register 相关插件在执行register时自行处理。
	return s.Plugins.DoRegister(name, server, metadata)
}

// 真正实现注册函数.
func (s *Server) register(name string, server interface{}, isUseName bool) (string, error) {

	// 注册需要进行上锁.
	s.serviceMapMu.Lock()
	defer s.serviceMapMu.Unlock()
	svce := new(service)
	svce.serType = reflect.TypeOf(server)
	svce.serValue = reflect.ValueOf(server)
	sname := reflect.Indirect(reflect.ValueOf(server)).Type().Name()
	if isUseName {
		sname = name
	}

	if sname == "" {
		return sname, fmt.Errorf("go_rpc.RegisterErr: no service name for type")
	}
	// 判断name是否可导.
	if !isUseName && !isExported(sname) {
		return sname, fmt.Errorf("go_rpc.RegisterErr: server name no exported")
	}
	svce.name = sname

	// init method.
	svce.method = initMethods(svce.serType)
	s.serviceMap[svce.name] = svce

	return sname, nil
}

func initMethods(svct reflect.Type) map[string]*methodType {

	// TODO: 待确定.
	methods := make(map[string]*methodType)
	for mi := 0; mi < svct.NumMethod(); mi++ {
		// 获取method和method.Type以及method.Name
		method := svct.Method(mi)
		mt := method.Type
		mname := method.Name
		// method必须可导出.
		//if method.PkgPath == "" {
		//	continue
		//}
		if mt.NumIn() != rpcMethodTotalArgs {
			continue
		}

		// 第0个参数一般是结构体本省.
		ctxType := mt.In(serverMethodContentArgsIndex)
		if !ctxType.Implements(typeOfContext) {
			continue
		}
		// 第二个参数判断是不是可导出的.
		argsType := mt.In(serverMethodReqArgsIndex)
		if !isExportedOrBuiltinType(argsType) {
			continue
		}
		// 第三个参数判断是不是指针.
		replyType := mt.In(serverMethodRspArgsIndex)
		if replyType.Kind() != reflect.Pointer {
			continue
		}

		// 方法返回值必须是ERROR.
		if returnType := mt.Out(0); returnType != typeOfError {
			// 方法返回值不是error则报错.
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argsType, replyType: replyType}
		// TODO: 可以将反射进行缓存起来.
	}

	// service
	return methods
}
