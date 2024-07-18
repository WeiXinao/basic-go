package main

const External = "包外"
const Internal = "包内"
const (
	a = 123
)

const (
	StatusA = iota
	StatusB
	StatusC
	StatusD

	StatusE = 100
	StatusF
)

const (
	Init = iota
	Running
	Paused
	Stop
)

const (
	DayA = iota*12 + 13
	DayB
)

func main() {
	const a = 123
	//a = 456
}
