// Package go_rpc.
// creator 2022-03-27 09:20:25
// Author  zhenxinma.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xmopen/gorpc/pkg/client"
	"github.com/xmopen/gorpc/pkg/server"
)

func main() {
	// test git add.1
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
		fmt.Println(err.Error())
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

// Add return a err test
func (r *RPCServer) Add(c context.Context, args *Args, res *Response) error {
	return fmt.Errorf("add err")
	//res.Result = "hello word" + args.Name
	//res.Age = 10 * args.Age
	//if res.Data.List == nil {
	//	res.Data.List = make([]int, 0, 10)
	//}
	//for i := 0; i < 10; i++ {
	//	res.Data.List = append(res.Data.List, i)
	//}
	//time.Sleep(2 * time.Second)
	//return nil
}
