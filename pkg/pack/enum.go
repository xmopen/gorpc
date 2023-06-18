package pack

// enum 定义Message编解码需要使用到的常量.

const (
	MagicNumber byte = 0b1001 // 魔数:9.

	// Version byte = 0b1001 // 版本:9.

	RequestTypeOfPeer2PeerRPC byte = 0b1 // 当前请求类型为点对点RPC.

	RequestNoOneWay byte = 0b0 // 当前请求需要进行响应.
	RequestIsOneWay byte = 0b1 // 当前请求需不需要进行响应.
)
