package main

import (
	"fmt"
	"github.com/funkygao/fae/servant/gen-go/fun/rpc"
	"github.com/funkygao/fae/servant/proxy"
	"time"
)

func main() {
	t1 := time.Now()

	client, err := proxy.New(5, 0).Servant(":9001")
	if err != nil {
		panic(err)
	}
	defer client.Recycle()

	ctx := rpc.NewContext()
	ctx.Caller = "me"
	for i := 0; i < 10; i++ {
		r, _ := client.Ping(ctx)

		fmt.Println(r, time.Since(t1))
		t1 = time.Now()
	}

}
