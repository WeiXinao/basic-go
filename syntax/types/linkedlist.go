package main

import "time"

type LinkedList struct {
	head *node
	tail *node

	// 这个就是公开熟悉，包外可访问
	Len int

	CreateTime time.Time
}

func (l *LinkedList) Add(idx int, val any) error {
	//TODO implement me
	panic("implement me")
}

func (l *LinkedList) Append(val any) {
	//TODO implement me
	panic("implement me")
}

func (l *LinkedList) Delete(idx int) (any, error) {
	//TODO implement me
	panic("implement me")
}

// 方法接收器 receiver
func (l *LinkedList) AddV1() {

}

// 编译时，需要确认结构体有多大
type node struct {
	prev *node
	next *node
}

type Integer int

func UseInt() {
	i1 := 10
	i2 := Integer(i1)
	var i3 Integer = 11
	println(i2, i3)
}
