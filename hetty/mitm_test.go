package hetty

import (
	"context"
	"github.com/jweny/go-mitm/hetty/xproxy"
	"testing"
	"time"
)

func TestProxy(t *testing.T) {
	serv, err := xproxy.NewServ(":8080")
	if err != nil {
		panic(err)
	}
	serv.Listen(context.Background())
	// 只是演示监听端口可以手动释放
	time.Sleep(120 * time.Second)
	err = serv.Stop()
	if err != nil {
		panic(err)
	}
}
