// Package test.
// creator 2022-03-30 13:54:29
// Author  zhenxinma.
package main

import (
	"fmt"
	"log"
	"net"
	"reflect"
	"time"
)

func main() {

	stu := &Student{
		Name: "zhangsan",
	}

	r := &Result{}

	// 反射不是很明白.
	//sv := reflect.ValueOf(stu).Elem()
	tv := reflect.TypeOf(stu)

	for i := 0; i < tv.NumMethod(); i++ {
		//method := sv.Method(i).Type()
		method := tv.Method(i)
		n := "lisi"
		rv := reflect.ValueOf(r).Interface()
		tv := reflect.TypeOf(r)
		// type的type？
		// **Student 什么鬼?
		t := reflect.New(tv).Interface()
		method.Func.Call([]reflect.Value{reflect.ValueOf(stu), reflect.ValueOf(n), reflect.ValueOf(t)})
		fmt.Println(rv)
	}
	fmt.Println("ok")
}

type Student struct {
	Name string
}

type Result struct {
	Name string
}

func (s *Student) GetName(name string, res *Result) string {
	res.Name = name + s.Name
	return s.Name + name
}

func GetStu() interface{} {
	return new(Student)
}

func testTcpServer() {

	// server.
	go func() {

		ln, err := net.Listen("tcp", "")
		if err != nil {
			log.Fatalln(err)
		}
		for {

			conn, err := ln.Accept()
			if err != nil {
				log.Fatalln(err)
			}
			// 1. 先读取20个字节.
			// 2. 剩下的根据字节数去读取.
			req := make([]byte, 2)
			n, err := conn.Read(req)
			if err != nil {
				log.Fatalln(err)
			}
			req = req[:n]
			log.Printf("length:%d result:%v", n, string(req))
		}
	}()

	time.Sleep(1 * time.Second)
	go func() {
		conn, err := net.Dial("tcp", "")
		if err != nil {
			log.Fatalln(err)
		}
		n, err := conn.Write([]byte("hello word"))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("client length:%d\n", n)
	}()
	time.Sleep(3 * time.Second)
}
