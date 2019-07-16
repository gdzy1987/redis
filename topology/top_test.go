package topology

import (
	"io"
	"testing"
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
	if len(tmp) > 1 {
		t.Fatal("Expected only 1 master")
	}

	for k, v := range tmp {
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
		t.Fatal("unmarshal byte cannot convert to Topologyer interface error")
	}

	if (nrs.UUID != rs.UUID) || (nrs.Master().Id != rs.Master().Id) {
		t.Fatal("unmarshal byte convert to Topologyer and soruce object value not equal")
	}
}

func TestSentinelMode(t *testing.T) {
	addrs, cancels := threeNode()
	defer cancels()

	rs := CreateRedisSentinel("", addrs...)
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
	if len(tmp) > 1 {
		t.Fatal("Expected only 1 master")
	}

	for k, v := range tmp {
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
		t.Fatal("unmarshal byte cannot convert to Topologyer interface error")
	}

	if (nrs.UUID != rs.UUID) || (nrs.Master().Id != rs.Master().Id) {
		t.Fatal("unmarshal byte convert to Topologyer and soruce object value not equal")
	}
}

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
		t.Fatal("unmarshal byte cannot convert to Topologyer interface error")
	}

	if len(ncs.Topology()) != len(rs.Topology()) {
		t.Fatal("unmarshal byte convert to Topologyer and soruce object value not equal")
	}
}
