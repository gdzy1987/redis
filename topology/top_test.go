package topology

import "testing"

func TestSingleMode(t *testing.T) {
	addrs, cancels := threeNode()
	defer cancels()

	rs := CreateRedisSingle("", addrs...)
	stop := rs.Run()
	defer stop()
	x := rs.MasterNodeInfo()
	if x[0].Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
		t.Fatal("The expected master id is inconsistent with the actual")
	}

	slaves := rs.SlaveNodeGroupInfo()

	for i := range slaves {
		switch slaves[i].Id {
		case "c41a2899e5c155a02a38e26450a916d1466379eb":
		case "c41a2899e5c155a02a38e26450a916d1466379ec":
		default:
			t.Fatal("Expectd Id not exists")
		}
	}
}

func TestSentinelMode(t *testing.T) {
	addrs, cancels := threeNode()
	defer cancels()

	rs := CreateRedisSentinel("", addrs...)
	stop := rs.Run()
	defer stop()
	x := rs.MasterNodeInfo()
	if x[0].Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
		t.Fatal("The expected master id is inconsistent with the actual")
	}

	slaves := rs.SlaveNodeGroupInfo()

	for i := range slaves {
		switch slaves[i].Id {
		case "c41a2899e5c155a02a38e26450a916d1466379eb":
		case "c41a2899e5c155a02a38e26450a916d1466379ec":
		default:
			t.Fatal("Expectd Id not exists")
		}
	}
}

func TestClusterMode(t *testing.T) {
	addrs, cancels := sixNode()
	defer cancels()
	rs := CreateRedisCluster("", addrs)
	stop := rs.Run()
	defer stop()
	xs := rs.MasterNodeInfo()

	for index := range xs {
		switch xs[index].Id {
		case "c41a2899e5c155a02a38e26450a916d1466379ea":
		case "c41a2899e5c155a02a38e26450a916d1466379ed":
		default:
			t.Fatal("The expected master id is inconsistent with the actual")
		}
	}
}
