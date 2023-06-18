package netpoll

import (
	"fmt"
	"testing"
)

func TestQueue(t *testing.T) {
	q := newQueue(4)
	q.Push(&Conn{
		state: 1,
	})
	q.Push(&Conn{
		state: 2,
	})
	q.Push(&Conn{
		state: 3,
	})
	q.Push(&Conn{
		state: 4,
	})
	for i := 0; i < int(q.size)+2; i++ {
		c := q.Get()
		fmt.Println(c.state)
	}
	//fmt.Println("---------------------------------------")
	//q.Delete()
	//for i := 0; i < int(q.size)+2; i++ {
	//	c := q.Get()
	//	fmt.Println(c.state)
	//}
}

func TestName(t *testing.T) {
	var uid uint8 = 0x03
	for i := 0; i < 20; i++ {
		uid = uid << 1
		fmt.Println(uid)
	}
}
