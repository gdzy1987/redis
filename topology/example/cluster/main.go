package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

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
	cluster := topology.CreateRedisCluster("",
		[][]string{
			[]string{"10.1.181.241:8001", "10.1.181.241:8006"},
			[]string{"10.1.181.241:8002", "10.1.181.241:8004"},
			[]string{"10.1.181.241:8003", "10.1.181.241:8005"},
		},
	)

	stop := cluster.Run()
	defer stop()
	top := cluster.Topology()
	if len(top) != 3 {
		panic("expected 3 gourp")
	}
}
