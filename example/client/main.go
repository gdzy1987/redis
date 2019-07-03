package main

import (
	"fmt"
	"time"

	client "github.com/dengzitong/redis/client"
)

func main() {
	cli := client.NewClient("127.0.0.1:6379", client.DialReadTimeout(time.Second*5))
	defer cli.Close()

	if reply, err := cli.Do("set", "x", "456"); err != nil {
		panic(err)
	} else if reply != client.OkReply {
		panic("reply error")
	}

	if reply, err := cli.Do("get", "_DSADSA_!"); err != nil {
		panic(err)
	} else if reply != client.NilReply {
		panic(fmt.Sprintf("reply error %s", reply))
	}

	// or password
	cli1 := client.NewClient("10.1.1.228:8003", client.DialReadTimeout(time.Second*5), client.DialPassword("wtf"))
	defer cli1.Close()

	if reply, err := cli1.Do("set", "x", "456"); err != nil {
		panic(err)
	} else if reply != client.OkReply {
		panic("reply error")
	}

	if reply, err := cli1.Do("get", "_DSADSA_!"); err != nil {
		panic(err)
	} else if reply != client.NilReply {
		panic(fmt.Sprintf("reply error %s", reply))
	}

	sentinelCli := client.NewClient("10.1.1.228:21001", client.DialReadTimeout(time.Second*5))
	defer sentinelCli.Close()

	if reply, err := client.Strings(sentinelCli.Do("sentinel", "master", "mymaster")); err != nil {
		panic(err)
	} else {
		// fmt.Printf("master reply: %v ", reply)
		_ = reply
	}

	if replys, err := client.MultiBulk(sentinelCli.Do("sentinel", "slaves", "mymaster")); err != nil {
		panic(err)
	} else {
		for _, reply := range replys {
			ss, err := client.Strings(reply, nil)
			if err != nil {
				panic(err)
			}
			fmt.Printf("slaves reply: %+s\n", ss)
		}
	}

}
