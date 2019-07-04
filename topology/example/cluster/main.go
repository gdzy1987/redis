package main

import (
	"github.com/dengzitong/redis/topology"
)

func main() {

	cluster := topology.CreateRedisClusterTop("10.1.1.228:21001", "10.1.1.228:21002", "10.1.1.228:21003")

	
}
