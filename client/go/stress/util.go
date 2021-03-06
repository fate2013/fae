package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strings"
)

func sampling(rate int) bool {
	return rand.Intn(rate) == 1
}

func showUsage() {
	flag.PrintDefaults()

	fmt.Println()
	fmt.Println("bitwise rpc calls")
	fmt.Println(strings.Repeat("=", 20))
	fmt.Printf("%16s %3d\n", "CallPing", CallPing)
	fmt.Printf("%16s %3d\n", "CallIdGen", CallIdGen)
	fmt.Printf("%16s %3d\n", "CallLCache", CallLCache)
	fmt.Printf("%16s %3d\n", "CallMemcache", CallMemcache)
	fmt.Printf("%16s %3d\n", "CallMongo", CallMongo)
	fmt.Printf("%16s %3d\n", "CallKvdb", CallKvdb)
	fmt.Printf("%16s %3d\n", "Ping+Idgen", CallPingIdgen)
	fmt.Printf("%16s %3d\n", "Lcache+Idgen", CallIdgenLcache)
}
