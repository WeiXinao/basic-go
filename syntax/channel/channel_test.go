package channel

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {
	//	这里是声明了一个放 int 类型的channel
	//	声明了但没有初始化，你读写这个都会崩溃
	//var ch chan int
	//ch <- 123
	//val := <-ch
	//println(val)

	//	放空结构体，一般用来做信号
	//	var chv1 chan struct {}

	//	不带容量要更加小心一些
	//	ch1 := make(chan int)
	//	这种就带容量的，容量是规定的，不会变的
	ch2 := make(chan int, 2)
	ch2 <- 123
	// 首先，你不能再发了
	// 但是你还可以读
	close(ch2)
	val, ok := <-ch2
	if !ok {
		//	ch2 已经被人关了
	}
	println(val)
}

func TestChangeClose(t *testing.T) {
	ch := make(chan int, 2)
	ch <- 123
	ch <- 234
	val, ok := <-ch
	t.Log(val, ok)
	close(ch)
	//close(ch)

	//	能不能把 234 读出来？
	val, ok = <-ch
	t.Log(val, ok)

	val, ok = <-ch
	t.Log(val, ok)
}

func SafeClose(ch chan int) {
	_, ok := <-ch
	if ok {
		close(ch)
	}
}

// 这个 ch 一定是 MyStruct 相关
type MyStruct struct {
	ch        chan struct{}
	closeOnce sync.Once
}

// 用户会多次调用，或者多个 goroutine 调用
func (m *MyStruct) Close() error {
	m.closeOnce.Do(func() {
		//	确保整个代码只会执行一次
		close(m.ch)
	})
	return nil
}

//func (m *MyStruct) UseV1(ch chan struct{}) error {
//	UseV2(ch)
//	UseV3(ch)
//}

type ChUsage struct {
}

type MyStructV1 struct {
	// 暴露出去了，你就不知道用户啥时候回给你关了
	Ch        chan struct{}
	closeOnce sync.Once
}

func TestLoopChannel(t *testing.T) {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
			time.Sleep(time.Millisecond * 100)
		}
		close(ch)
	}()
	for val := range ch {
		t.Log(val)
	}

	t.Log("channel 被关了")
}

func TestChannelBlock(t *testing.T) {
	ch := make(chan int, 3)
	//	阻塞，不再占用 CPU 了
	val := <-ch
	//	意味着，这一句不会执行下来
	t.Log(val)
	//	goroutine 数量
	runtime.NumGoroutine()
}

func TestGoroutineChRead(t *testing.T) {
	ch := make(chan int, 100000)
	// 这一个就泄露掉了
	def := new(BigObj)
	go func() {
		for i := 0; i < 100000; i++ {
			ch <- i
		}
		abc := new(BigObj)
		t.Log(abc)
		t.Log(def)
		//	永久阻塞在这里，ch 会占据内存，永远不会被回收
		ch <- 1
	}()

	//	这里后面没人往 ch 里面读数据
}

type BigObj struct {
}

func TestSelect(t *testing.T) {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)
	ch1 <- 123
	ch2 <- 234
	//go func() {
	//	time.Sleep(time.Millisecond * 100)
	//	ch1 <- 123
	//}()
	//go func() {
	//	time.Sleep(time.Millisecond * 100)
	//	ch2 <- 123
	//}()
	select {
	case val := <-ch1:
		t.Log("ch1", val)
		val = <-ch2
		t.Log("ch2", val)
	case val := <-ch2:
		t.Log("ch2", val)
		val = <-ch1
		t.Log("ch1", val)
	}
}
