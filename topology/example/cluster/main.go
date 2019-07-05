package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dengzitong/redis/topology"
)

func main() {
	log.SetOutput(os.Stdout)
	cluster := topology.CreateRedisClusterTopFromAddrs("10.1.1.228:7001", "10.1.1.228:7002", "10.1.1.228:7003")

	stop := cluster.Run()
	defer stop()

	i, d := cluster.ReceiveNodeInfos()
	for {
		select {
		case nodes, ok := <-i:
			if !ok {
				return
			}
			fmt.Printf("add node %#v\n", nodes)

		case nodes, ok := <-d:
			if !ok {
				return
			}
			fmt.Printf("delete node %#v\n", nodes)
		}
	}

}
