package main

import (
	"bytes"
	"encoding/json"
	"fmt"

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

	var toplogyist topology.Topologist
	toplogyist = topology.CreateRedisCluster("",
		[][]string{
			[]string{"10.1.181.241:8001", "10.1.181.241:8006"},
			[]string{"10.1.181.241:8002", "10.1.181.241:8004"},
			[]string{"10.1.181.241:8003", "10.1.181.241:8005"},
		},
	)

	stop := toplogyist.Run()
	defer stop()
	top := toplogyist.Topology()
	if len(top) != 3 {
		panic("expected 3 gourp")
	}

	for k, v := range top {
		println(jsonPrettyPrint(fmt.Sprintf(`{ master:"%+v" , slave:"%+v" }`, k, v[1])))
	}
}
