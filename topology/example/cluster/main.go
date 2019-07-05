package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dengzitong/redis/topology"
)

func main() {
	log.SetOutput(os.Stdout)
	cluster := topology.CreateRedisClusterTopFromAddrs("10.1.1.228:7001", "10.1.1.228:7002")

	stop := cluster.Run()
	defer stop()

	i, d := cluster.ReceiveNodeInfos()
	for {
		select {
		case nodes, ok := <-i:
			if !ok {
				return
			}
			for _, node := range nodes {
				fmt.Printf("add node %v\n", node)
			}

		case nodes, ok := <-d:
			if !ok {
				return
			}
			for _, node := range nodes {
				fmt.Printf("delete node %v\n", node)
			}
		}
	}

}
