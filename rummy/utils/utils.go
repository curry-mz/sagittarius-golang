package utils

import (
	"reflect"
	"time"
	"unsafe"
)

func ABS64(v int64) int64 {
	if v >= 0 {
		return v
	}
	return -v
}

func String2Bytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func DayZeroSec() int64 {
	dt := time.Unix(time.Now().Unix(), 0)
	dtz := time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, dt.Location())
	return dtz.Unix()
}
func SumSlice(slice []int64) int64 {
	var sum int64 = 0
	for _, num := range slice {
		sum += num
	}
	return sum
}
func DoubleSlice(slice []int64) []int64 {

	var result2 []int64
	var sum int64 = 0

	for _, num := range slice {
		sum += num
	}
	for _, num := range slice {
		result2 = append(result2, num)
	}
	slice = append(slice, result2...)
	return slice
}

// 查询某个字符串是否切片中
func InArrayString(s string, arr []string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}
