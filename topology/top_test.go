package topology

import (
	"io"
	"testing"

	"github.com/iauto360x/redis/client"
)

var _ = io.Writer(&dstTest{})

type dstTest struct {
	data []byte
}

func (d *dstTest) Write(p []byte) (n int, err error) {
	d.data = make([]byte, len(p), len(p))
	return copy(d.data, p), nil
}

func TestSingleMode(t *testing.T) {
	addrs, cancels := threeNode()
	defer cancels()

	rs := CreateRedisSingle("", addrs...)
	stop := rs.Run()
	defer stop()
	x := rs.masterNodeInfo()
	if x.Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
		t.Fatal("The expected master id is inconsistent with the actual")
	}

	slaves := rs.slaveNodeGroupInfo()

	for i := range slaves {
		switch slaves[i].Id {
		case "c41a2899e5c155a02a38e26450a916d1466379eb":
		case "c41a2899e5c155a02a38e26450a916d1466379ec":
		default:
			t.Fatal("Expectd Id not exists")
		}
	}

	tmp := rs.Topology()
	if len(*tmp) > 1 {
		t.Fatal("Expected only 1 master")
	}

	for k, v := range *tmp {
		if !k.IsMaster {
			t.Fatal("Expected key is master")
		}

		if k.Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
			t.Fatal("The expected master id is inconsistent with the actual")
		}

		if len(v) != 2 {
			t.Fatal("Expected only 2 slave")
		}

		for i := range v {
			switch v[i].Id {
			case "c41a2899e5c155a02a38e26450a916d1466379eb":
			case "c41a2899e5c155a02a38e26450a916d1466379ec":
			default:
				t.Fatal("Expectd Id not exists")
			}
		}

		_ = v
	}

	dstbuf := &dstTest{}
	err := rs.MarshalToWriter(dstbuf)
	if err != nil {
		t.Fatal(err)
	}

	th, err := UnmarshalFromBytes(SingleMode, dstbuf.data)
	if err != nil {
		t.Fatal(err)
	}

	nrs, ok := th.(*RedisSingle)
	if !ok {
		t.Fatal("unmarshal byte cannot convert to Topologist interface error")
	}

	if (nrs.UUID != rs.UUID) || (nrs.Master().Id != rs.Master().Id) {
		t.Fatal("unmarshal byte convert to Topologist and soruce object value not equal")
	}
}

// func TestSentinelMode(t *testing.T) {
// 	addrs, cancels := threeNode()
// 	defer cancels()

// 	rs := CreateRedisSentinel("", addrs...)
// 	stop := rs.Run()
// 	defer stop()
// 	top := rs.Topology()
// 	for k, slaves := range top {
// 		if k.Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
// 			t.Fatal("The expected master id is inconsistent with the actual")
// 		}

// 		for i := range slaves {
// 			switch slaves[i].Id {
// 			case "c41a2899e5c155a02a38e26450a916d1466379eb":
// 			case "c41a2899e5c155a02a38e26450a916d1466379ec":
// 			default:
// 				t.Fatal("Expectd Id not exists")
// 			}
// 		}
// 	}

// 	tmp := rs.Topology()
// 	if len(tmp) > 1 {
// 		t.Fatal("Expected only 1 master")
// 	}

// 	for k, v := range tmp {
// 		if !k.IsMaster {
// 			t.Fatal("Expected key is master")
// 		}

// 		if k.Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
// 			t.Fatal("The expected master id is inconsistent with the actual")
// 		}

// 		if len(v) != 2 {
// 			t.Fatal("Expected only 2 slave")
// 		}

// 		for i := range v {
// 			switch v[i].Id {
// 			case "c41a2899e5c155a02a38e26450a916d1466379eb":
// 			case "c41a2899e5c155a02a38e26450a916d1466379ec":
// 			default:
// 				t.Fatal("Expectd Id not exists")
// 			}
// 		}

// 		_ = v
// 	}

// 	dstbuf := &dstTest{}
// 	err := rs.MarshalToWriter(dstbuf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	th, err := UnmarshalFromBytes(SingleMode, dstbuf.data)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	nrs, ok := th.(*RedisSingle)
// 	if !ok {
// 		t.Fatal("unmarshal byte cannot convert to Topologist interface error")
// 	}

// 	if (nrs.UUID != rs.UUID) || (nrs.Master().Id != rs.Master().Id) {
// 		t.Fatal("unmarshal byte convert to Topologist and soruce object value not equal")
// 	}
// }

func TestClusterMode(t *testing.T) {
	addrs, cancels := sixNode()
	defer cancels()
	rs := CreateRedisCluster("", addrs)
	stop := rs.Run()
	defer stop()
	xs := rs.masterNodeInfo()

	for index := range xs {
		switch xs[index].Id {
		case "c41a2899e5c155a02a38e26450a916d1466379ea":
		case "c41a2899e5c155a02a38e26450a916d1466379ed":
		default:
			t.Fatal("The expected master id is inconsistent with the actual")
		}
	}

	dstbuf := &dstTest{}
	err := rs.MarshalToWriter(dstbuf)
	if err != nil {
		t.Fatal(err)
	}

	th, err := UnmarshalFromBytes(ClusterMode, dstbuf.data)
	if err != nil {
		t.Fatal(err)
	}

	ncs, ok := th.(*RedisCluster)
	if !ok {
		t.Fatal("unmarshal byte cannot convert to Topologist interface error")
	}

	if len(*(ncs.Topology())) != len(*(rs.Topology())) {
		t.Fatal("unmarshal byte convert to Topologist and soruce object value not equal")
	}
}

var (
	otherClusterNodeInfo1 = `cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected
`

	cnode1 = `# Server
redis_version:4.0.2
run_id:c6d165b72cfcd76d7662e559dc709e00e3dabf03
tcp_port:7001

# Replication
role:master
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=100,lag=1
master_repl_offset:100`

	cnode2 = `# Server
redis_version:4.0.2
run_id:656042ad560b887164138a19dab2502154f8b039
tcp_port:7004

# Replication
role:slave
connected_slaves:1
master_repl_offset:100`

	cnode3 = `# Server
redis_version:4.0.2
run_id:885493415bea22919fc9ce83836a9e6a8d0c1314
tcp_port:7003

# Replication
role:master
connected_slaves:1
slave0:ip=10.1.1.228,port=7006,state=online,offset=100,lag=1
master_repl_offset:100`

	cnode4 = `# Server
redis_version:4.0.2
run_id:62bd020a2a5121a27c0e5540d1f0d4bba08cebb2
tcp_port:7006

# Replication
role:slave
connected_slaves:1
master_repl_offset:100`

	cnode5 = `# Server
redis_version:4.0.2
run_id:cebd9205cbde0d1ec4ad75600849a88f1f6294f6
tcp_port:7005

# Replication
role:master
connected_slaves:1
slave0:ip=10.1.1.228,port=7002,state=online,offset=100,lag=1
master_repl_offset:100`

	cnode6 = `# Server
redis_version:4.0.2
run_id:a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab
tcp_port:7002

# Replication
role:slave
connected_slaves:1
master_repl_offset:100`
)

func clusterSixNode() ([][]string, func()) {
	s1 := client.NewVirtualServer(nil)
	s1.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node1))
	})
	s1.AddHandles("cluster", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(otherClusterNodeInfo1))
	})

	cancel1, s1addr := s1.Serve("localhost:0")

	s2 := client.NewVirtualServer(nil)
	s2.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node2))
	})
	s2.AddHandles("cluster", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(otherClusterNodeInfo1))
	})
	cancel2, s2addr := s2.Serve("localhost:0")

	s3 := client.NewVirtualServer(nil)
	s3.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node3))
	})
	s3.AddHandles("cluster", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(otherClusterNodeInfo1))
	})
	cancel3, s3addr := s3.Serve("localhost:0")

	s4 := client.NewVirtualServer(nil)
	s4.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node4))
	})
	s4.AddHandles("cluster", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(otherClusterNodeInfo1))
	})
	cancel4, s4addr := s4.Serve("localhost:0")

	s5 := client.NewVirtualServer(nil)
	s5.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node5))
	})
	s5.AddHandles("cluster", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(otherClusterNodeInfo1))
	})
	cancel5, s5addr := s5.Serve("localhost:0")

	s6 := client.NewVirtualServer(nil)
	s6.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node6))
	})
	s6.AddHandles("cluster", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(otherClusterNodeInfo1))
	})
	cancel6, s6addr := s6.Serve("localhost:0")

	return [][]string{[]string{s1addr, s2addr}, []string{s3addr, s4addr}, []string{s5addr, s6addr}}, func() {
		defer func() {
			for _, cancel := range []func(){
				cancel1,
				cancel2,
				cancel3,
				cancel4,
				cancel5,
				cancel6,
			} {
				cancel()
			}
		}()
	}

}

func TestClusterAddr(t *testing.T) {
	addrs, cancels := clusterSixNode()
	defer cancels()

	argsAddr := make([]string, 0)

	for i := range addrs {
		for y := range addrs[i] {
			argsAddr = append(argsAddr, addrs[i][y])
		}
	}
	// value error
	realAddrs, err := clusterAddr("", argsAddr...)
	if err != nil {
		t.Fatal(err)
	}

	if len(realAddrs) != 3 {
		t.Fatal("unkonw error")
	}
}
