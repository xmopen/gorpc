package util

import (
	"unsafe"
)

const (
	ContextKeyCustTime = "cust_time"
)

// CustTime  统计消耗时间.
type CustTime struct {
	ServiceName string
	MethodName  string
	Start       int64
}

// BytesSliceToString 快速的将字节数组转换成字符串.
func BytesSliceToString(data []byte) string {
	// 1、先将字节数组转换成指针.直接指向内存.
	return *(*string)(unsafe.Pointer(&data))
}

// StructToBytesSlice 将结构体转换成字节数组.
func StructToBytesSlice(data interface{}) (interface{}, bool) {
	if data == nil {
		return nil, false
	}
	if res, ok := data.([]byte); ok {
		return res, true
	}
	if res, ok := data.([]byte); ok {
		return res, true
	}
	return nil, false
}
