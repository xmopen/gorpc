# 一、使用Demo.

```go
// Package go_rpc.
// creator 2022-03-27 09:20:25
// Author  zhenxinma.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gitee.com/zhenxinma/go_rpc/pkg/client"
	"gitee.com/zhenxinma/go_rpc/pkg/server"
)

func main() {

	go func() {
		s := server.NewServer()
		s.SetTrace(true)
		s.RegisterName("RPCServer", new(RPCServer), "")
		if err := s.Server("tcp", ":8899"); err != nil {
			log.Fatalln(err)
		}
		fmt.Println("server starting...")
	}()

	time.Sleep(1 * time.Second)
	// 构造Client.
	c, err := client.NewClient("tcp", ":8899", nil)
	c.Trace = true
	if err != nil {
		log.Fatalln(err)
	}
	args := &Args{
		Age:  100,
		Name: "张三",
	}
	response := &Response{}
	err = c.Call(context.Background(), "RPCServer", "Add", args, response)
	if err != nil {
		// timeout.
		log.Fatalln(err)
	}
}

type Args struct {
	Name string
	Age  int64
}

type Response struct {
	Result string
	Age    int64
	Data   struct {
		List []int
	}
}

// RPCServer 服务.
type RPCServer struct {
}

func (r *RPCServer) Add(c context.Context, args *Args, res *Response) error {
	res.Result = "hello word" + args.Name
	res.Age = 10 * args.Age
	if res.Data.List == nil {
		res.Data.List = make([]int, 0, 10)
	}
	for i := 0; i < 10; i++ {
		res.Data.List = append(res.Data.List, i)
	}
	time.Sleep(2 * time.Second)
	return nil
}


//  输出结果.
2022-09-17 20:18:54.742[gorpc.client][INFO]servicename:RPCServer methodname:Add calltime:2s result:&{Result:hello word张三 Age:1000 Data:{List:[0 1 2 3 4 5 6 7 8 9]}}[D:/code/go/src/me.code/gorpc/pkg/client/client.go:98]

```

#  二、自定义协议

![image-20220917221154256](https://typoraimg-1303903194.cos.ap-guangzhou.myqcloud.com/image-20220917221154256.png)

**消息的组成部分：**

- 第一个字节：魔数
- 第二个字节：版本
- 第三个字节：请求类型
- 第四个字节：超时控制
- 第五个字节到第八个字节：payload length,请求体长度
- 第9个字节到第12个字节：data length 相应长度
- 第13个字节到第16个字节：请求服务名称长度
- 第17个字节到第20个字节：请求方法名称长度
- `sl+ml+pl+dl` 长度字节就是本次请求内容，依次为服务名称、方法名称、请求体、响应体。