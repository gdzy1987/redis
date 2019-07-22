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
	if len(*top) != 3 {
		panic("expected 3 gourp")
	}
	for k, v := range *top {
		toplogyist.Increment(k, 12345)
		println(jsonPrettyPrint(fmt.Sprintf(`{ master:"%+v" , slave:"%+v" }`, k, v[1])))
		s := toplogyist.Offset(k)
		if s != "12345" {
			panic("expected value is unusual")
		}
		// group := toplogyist.Group(k)
		// group.CloseAllMember()
	}
	stop()

	// SentinelMode Redis server
	toplogyist = topology.CreateRedisSentinel("wtf", "10.1.1.228:21001", "10.1.1.228:21002", "10.1.1.228:21003")
	stop1 := toplogyist.Run()
	top = toplogyist.Topology()
	for k, v := range *top {
		toplogyist.Increment(nil, 12345)
		println(jsonPrettyPrint(fmt.Sprintf(`{ master:"%+v" , slaves:"%#v" }`, k, v)))
		s := toplogyist.Offset(k)
		if s != "12345" {
			panic("expected value is unusual")
		}
	}
	stop1()

	// SingleMode Redis server
	toplogyist = topology.CreateRedisSingle("", "127.0.0.1:6379")
	stop2 := toplogyist.Run()
	top = toplogyist.Topology()
	for k, v := range *top {
		toplogyist.Increment(nil, 12345)
		println(jsonPrettyPrint(fmt.Sprintf(`{ master:"%+v" , slaves:"%#v" }`, k, v)))
		s := toplogyist.Offset(k)
		if s != "12345" {
			panic("expected value is unusual")
		}
	}

	top1 := toplogyist.Topology()
	n, o, h := top.Compares(top1)
	if h {
		panic("beyond expectation")
	}
	_, _ = n, o

	stop2()
}
