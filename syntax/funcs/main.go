package main

func main() {
	//name, age := Func10()
	//println(name, age)
	//
	//name1, _ := Func10()
	//println(name1)
	//// 使用 := 前提，就是左边必有至少一个新变量
	//name1, name2 := Func10()
	//println(name1, name2)
	//
	//Func6("Hello", "Xin")

	//Recursive(10)

	//UseFunctional4()
	//functional8()

	//fn := Closure("小新")
	// fn 其实已经从 Closure 里面返回了
	// 但是我 fn 还要用到 "小新"
	//println(fn())

	//getAge := Closure1()
	//println(getAge())
	//println(getAge())
	//println(getAge())
	//println(getAge())
	//println(getAge())
	//
	//getAge = Closure1()
	//println(getAge())
	//println(getAge())
	//println(getAge())
	//println(getAge())
	//println(getAge())

	//Defer()
	//DeferClosure()
	//DeferClosureV1()

	println(DeferReturn())
	println(DeferReturnV1())
	DeferFuncSequence()
	DeferClosureLoopV1()
	DeferClosureLoopV2()
	DeferClosureLoopV3()
}
