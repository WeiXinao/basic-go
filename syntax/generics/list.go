package main

// T 类型参数，名字叫做 T，约束是 any，等于没有约束
type List[T any] interface {
	Add(idx int, t T)
	Append(t T)
}

func main() {
	//UseList()
	println(Sum[int](1, 2, 3))
	println(Sum[Integer](1, 2, 3))
	println(Sum[float64](1.1, 2.1, 3.1))
	println(Sum[float32](1.1, 2.1, 3.1))
	var j MyMarshal
	ReleaseResource[*MyMarshal](&j)
}

type MyMarshal struct {
}

func (m *MyMarshal) MarshalJSON() ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func UseList() {
	//var l List[int]
	//l.Append(12)

	var lany List[any]
	lany.Append(12.3)
	lany.Append(123)
	lk := LinkedList[int]{}
	intVal := lk.head.val
	println(intVal)
}

// type parameter
type LinkedList[Element any] struct {
	head *node[Element]
	t    Element
	tp   *Element
	tp2  ******Element
}

type node[T any] struct {
	val T
}
