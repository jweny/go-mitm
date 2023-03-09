package gproxy

import (
	"context"
	"testing"
	"time"
)

func TestNewProxy(t *testing.T) {
	proxy := NewProxy("127.0.0.1:8080")
	err := setCA()
	if err != nil {
		panic(err)
	}
	proxy.Listen(context.Background())
	// 只是演示监听端口可以手动释放
	time.Sleep(120 * time.Second)
	err = proxy.Stop()

	if err != nil {
		panic(err)
	}
}
