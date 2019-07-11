package topology

import "testing"

func TestSingleMode(t *testing.T) {
	addrs, cancels := threeNode()
	defer cancels()

	rs := CreateRedisSingle("", addrs...)
	stop := rs.Run()
	defer stop()
	x := <-rs.ReceiveNodeInfos()
	if x[0].Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
		t.Fatal("The expected master id is inconsistent with the actual")
	}
}
