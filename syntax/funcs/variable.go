package main

func YourName(name string, aliases ...string) {
	//alias 是一个切片
}

func CallYourName() {
	YourName("小新")
	YourName("孙悟空", "卡卡罗特")
	aliases := []string{"孙悟空", "卡卡罗特"}
	YourName("悟空", aliases...)
}
