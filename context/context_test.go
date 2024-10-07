package context

import (
	"context"
	"testing"
	"time"
)

type Key1 struct {
}

func TestContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), Key1{}, "value1")
	val := ctx.Value(Key1{})
	t.Log(val)
	ctx = context.WithValue(ctx, "key2", "value2")
	val = ctx.Value(Key1{})
	t.Log(val)
	ctx = context.WithValue(ctx, Key1{}, "value1-1")
	val = ctx.Value(Key1{})
	t.Log(val)
}

func TestContext_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	//	为什么一定要 cancel 呢？
	//	防止 goroutine 泄露
	cancel()

	//	防止有些人使用了 Done, 在等待 ctx 结束信号
	go func() {
		ch := ctx.Done()
		<-ch
	}()

	//	在这里用 ctx

	ctx = context.WithValue(ctx, Key1{}, "value1-1")
	val := ctx.Value(Key1{})
	t.Log(val)

	ctx, cancel = context.WithTimeout(ctx, time.Second)
	cancel()
	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(time.Second))
	cancel()
}

func TestContextErr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()
	t.Log(ctx)
	//	你怎么区别被取消了，还是超时了呢？
	if ctx.Err() == context.Canceled {

	} else if ctx.Err() == context.DeadlineExceeded {

	}
}

func TestContextSub(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	_, cancel1 := context.WithCancel(ctx)
	go func() {
		time.Sleep(time.Second)
		cancel1()
	}()

	go func() {
		// 监听 subCtx 结束的信号
		t.Log("等待信号...")
		<-ctx.Done()
		t.Log("收到信号...")
	}()
	time.Sleep(time.Second * 10)
}

//func MockIO() {
//	select {
//	// 监听超时
//	case <-ctx.Done():
//	case <-biz.Signal():
//		//	监听你的正常业务
//	}
//}
