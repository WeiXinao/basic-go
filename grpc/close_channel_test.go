package grpc

import (
	"testing"
	"time"
)

func TestCloseChannel(t *testing.T) {
	var closeCh = make(chan struct{})
	go func() {
		time.Sleep(time.Second)
		close(closeCh)
	}()

	go func() {
		for {
			select {
			case <-closeCh:
				t.Log("closeCh closed")
				return
			}
		}
	}()

	time.Sleep(5 * time.Second)
}
