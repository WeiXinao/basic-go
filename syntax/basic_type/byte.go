package main

import (
	"fmt"
	"unsafe"
)

func Byte() {
	var a byte = 'a'
	println(a)
	println(fmt.Sprintf("%c", a))

	var str string = "this is string"
	var bs []byte = []byte(str) // 转化时发生了复制
	bs[0] = 'T'
	println(str)

	bsp := (*[2]uintptr)(unsafe.Pointer(&str))
	bspt := [3]uintptr{bsp[0], bsp[1], bsp[1]}
	bs2 := *(*[]byte)(unsafe.Pointer(&bspt))
	fmt.Printf("%s", bs2)
	//bs2[0] = 'T'
	//fmt.Printf("%s", str)
}
