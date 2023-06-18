package errcode

import "fmt"

// err中定义Error类型.

// Client.

var (
	// ClientCallTimeOut 客户端调用超时.
	ClientCallTimeOut = fmt.Errorf("gorpc client call timeout")
)

var (
	CheckMagicError   = fmt.Errorf("gorpc check magic err")
	CheckVersionError = fmt.Errorf("gorpc check version err")

	// SendResponseLengthNotFull client write 写入到tcp conn中的字节长度和本次实际要发送的字节长度不匹配.
	SendResponseLengthNotFull = fmt.Errorf("tcp write length not is data length")

	// TcpConnWriteLengthNotFull 写入到net.conn中的长度不是预期长度.
	TcpConnWriteLengthNotFull = fmt.Errorf("tcp write length not is data length")

	// ClientCallPanicError 客户端远程调用时发生panic,防止程序崩溃,捕获panic并且用xlog输出.
	ClientCallPanicError = fmt.Errorf("gorpc client call panic error")

	// ServerMethodResponseNotPointer 服务方法返回值接受类型不是指针.
	ServerMethodResponseNotPointer = fmt.Errorf("gorpc server response not is poninter")
)
