package main

import "fmt"

func Closure(name string) func() string {
	return func() string {
		return "hello, " + name
	}
}

func Closure1() func() int {
	var age = 0
	fmt.Printf("out: %p\n", &age)
	return func() int {
		age = age + 1
		fmt.Printf("%p\n", &age)
		return age
	}
}
