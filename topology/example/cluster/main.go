package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dengzitong/redis/topology"
)

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

func main() {
	log.SetOutput(os.Stdout)
	cluster := topology.CreateRedisClusterTopFromAddrs("",
		"10.1.181.241:8001",
		"10.1.181.241:8002",
		"10.1.181.241:8003",
		"10.1.181.241:8004",
		"10.1.181.241:8005",
		"10.1.181.241:8006",
	)

	stop := cluster.Run()
	defer stop()

	i, d := cluster.ReceiveNodeInfos()

	ticker := time.NewTicker(10 * time.Second)
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
		case <-ticker.C:
			fmt.Printf("%s\n", jsonPrettyPrint(cluster.Format()))
			fmt.Printf("%s\n", strings.Repeat("-", 100))
		}
	}

}
