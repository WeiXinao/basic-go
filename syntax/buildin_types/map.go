package main

func Map() {
	m1 := map[string]int{
		"key1": 123,
	}
	m1["hello"] = 345

	// 容量
	m2 := make(map[string]int, 12)
	m2["key2"] = 12

	val, ok := m1["小新"]
	if ok {
		// 有这个键值对
		println(val)
	}

	val = m1["大新"]
	println("大新对应的值：", val)

	delete(m1, "key3")
}
