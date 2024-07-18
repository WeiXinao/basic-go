package main

import (
	"unicode/utf8"
)

func String() {
	// He said: "Hello Go!"
	println("He said: \"Hello Go!\"")
	println(`
可以换行
再一行
`)
	println("hello" + "go")
	//println("hello" + string(123))
	//print(fmt.Sprintf("hello %d", 123))
	println(len("abc"))

	println(len("你好"))
	println(len("你好abc"))
	println(utf8.RuneCountInString("你好"))
	//strings.Cut()
}
