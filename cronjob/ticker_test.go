package cronjob

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	ticker := time.NewTicker(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	defer ticker.Stop()
	//	每隔一秒钟就会有一个信号
end:
	for {
		select {
		case <-ctx.Done():
			//	循环结束
			t.Log("循环结束")
			break end
		case now := <-ticker.C:
			t.Log("过了一秒", now.UnixMilli())
		}
	}
	t.Log("结束程序")
}
