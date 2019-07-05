package topology

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func listener(addrs ...string) []net.Listener {
	listeners := make([]net.Listener, 0)
	for _, addr := range addrs {
		listen, err := net.Listen("tcp", addr)
		if err != nil {
			//
		}
		listeners = append(listeners, listen)
	}
	return listeners
}

func TestPing(t *testing.T) {
	addrs := []string{
		"127.0.0.1:50001",
		"127.0.0.1:50002",
		"127.0.0.1:50003",
		"127.0.0.1:2181",
		"127.0.0.1:2180",
	}

	lns := listener(addrs...)
	defer func() {
		for _, ln := range lns {
			ln.Close()
		}
	}()

	t.Run("ping", func(t *testing.T) {
		_addrs, err := Ping(addrs...)
		if err != nil {
			t.Fatal(err)
		}
		if len(_addrs) != len(lns) {
			t.Fatal("expected addrs length not equal")
		}
	})
}

func TestCommandSize(t *testing.T) {
	b, size := ConvertAcmdSize("ping")
	if size != 14 {
		t.Fatal("expected length not equal.")
	}
	if !bytes.Equal(b, []byte("*1\r\n$4\r\nping\r\n")) {
		t.Fatal("expected value not equal.")
	}
}

var (
	ClusterNodeInfo = `cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected
`
	SentinelNodeInfo = []string{
		"name",
		"mymaster",
		"ip",
		"10.1.1.228",
		"port",
		"8003",
		"runid",
		"8be9401d095b9072b2e67c59fea68894e14da193",
		"flags",
		"master",
		"link-pending-commands",
		"0",
		"link-refcount",
		"1",
		"last-ping-sent",
		"0",
		"last-ok-ping-reply",
		"848",
		"last-ping-reply",
		"848",
		"down-after-milliseconds",
		"60000",
		"info-refresh",
		"10000",
		"role-reported",
		"master",
		"role-reported-time",
		"222861310",
		"config-epoch",
		"4",
		"num-slaves",
		"124",
		"num-other-sentinels",
		"2",
		"quorum",
		"2",
		"failover-timeout",
		"180000",
		"parallel-syncs",
		"1",
	}
)

func TestProbeTopParsedByInfo(t *testing.T) {

	clusterAddrs, err := parsedByInfo(ClusterMode, ClusterNodeInfo)
	if err != nil {
		t.Fatal(err)
	} else if len(clusterAddrs) < 1 {
		t.Fatal("need at least 1 master node")
	}

	sentinelAddrs, err := parsedByInfo(SentinelMode, SentinelNodeInfo)
	if err != nil {
		t.Fatal(err)
	} else if len(sentinelAddrs) < 1 {
		t.Fatal("need at least 1 master node")
	}

	singleAddrs, err := parsedByInfo(SingleMode, "localhost:1234")
	if err != nil {
		t.Fatal(err)
	} else if len(singleAddrs) < 1 {
		t.Fatal("need at least 1 master node")
	}
}

func TestExec(t *testing.T) {
	err := ExecTimeout(time.Second*1, "ps", "-ef")
	if err != nil {
		t.Fatal(err)
	}
}

func TestProbeFunc(t *testing.T) {
	ss, err := ProbeTopology(ClusterMode, "10.1.1.228:7001")
	if err != nil {
		t.Fatal(err)
	} else if len(ss) < 1 {
		t.Fatal("expected cluster node")
	}
}
