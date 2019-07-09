package client

import (
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

var (
	b2s = func(p []byte) string { return *(*string)(unsafe.Pointer(&p)) }
	s2b = func(s string) []byte { return *(*[]byte)(unsafe.Pointer(&s)) }
)

type mm struct {
	mu   sync.Mutex
	data map[string]string
}

func (m *mm) Set(key []byte, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[b2s(key)] = b2s(value)
	return nil
}

func (m *mm) Get(k []byte) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exist := m.data[b2s(k)]; !exist {
		return nil, ErrNil
	}
	return s2b(m.data[b2s(k)]), nil
}

func (m *mm) Hset(k []byte, f []byte, v []byte) error {
	return nil
}
func (m *mm) Hget([]byte, []byte) ([]byte, error) {
	return nil, nil
}

func (m *mm) Cluster() ([]byte, error) {
	return s2b(otherClusterNodeInfo), nil
}

func (m *mm) Sentinel() ([]interface{}, error) {
	return SentinelNodeInfo, nil
}

var (
	ErrCommand           = `(error) ERR wrong number of arguments for '%s' command`
	Nil                  = `(nil)`
	Ok                   = `OK`
	otherClusterNodeInfo = `cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected`
	SentinelNodeInfo = []interface{}{
		[]byte("name"),
		[]byte("mymaster"),
		[]byte("ip"),
		[]byte("10.1.1.228"),
		[]byte("port"),
		[]byte("8003"),
		[]byte("runid"),
		[]byte("8be9401d095b9072b2e67c59fea68894e14da193"),
		[]byte("flags"),
		[]byte("master"),
		[]byte("link-pending-commands"),
		[]byte("0"),
		[]byte("link-refcount"),
		[]byte("1"),
		[]byte("last-ping-sent"),
		[]byte("0"),
		[]byte("last-ok-ping-reply"),
		[]byte("848"),
		[]byte("last-ping-reply"),
		[]byte("848"),
		[]byte("down-after-milliseconds"),
		[]byte("60000"),
		[]byte("info-refresh"),
		[]byte("10000"),
		[]byte("role-reported"),
		[]byte("master"),
		[]byte("role-reported-time"),
		[]byte("222861310"),
		[]byte("config-epoch"),
		[]byte("4"),
		[]byte("num-slaves"),
		[]byte("124"),
		[]byte("num-other-sentinels"),
		[]byte("2"),
		[]byte("quorum"),
		[]byte("2"),
		[]byte("failover-timeout"),
		[]byte("180000"),
		[]byte("parallel-syncs"),
		[]byte("1"),
	}
	stge = &mm{
		mu:   sync.Mutex{},
		data: make(map[string]string),
	}
)

func localAddr(s string) string {
	if strings.Contains(s, "[::]") {
		// [::]:12345
		ss := strings.Split(s, ":")
		return ss[3]
	}
	return strings.Split(s, ":")[1]
}

func TestClientPing(t *testing.T) {

	virtualServer := NewVirtualServer(stge)

	cancel, addr := virtualServer.Serve("localhost:0")
	defer cancel()

	virtualServer.AddHandles("ping", func(rw *RespWriter, req [][]byte) error {
		return rw.FlushString("PONG")
	})

	cli := NewClient("localhost:" + localAddr(addr))
	defer cli.Close()

	if reply, err := String(cli.Do("PING")); err != nil {
		t.Fatal(err)
	} else if reply != "PONG" {
		t.Fatal("Expected value does not match")
	}

}

func TestClientSet(t *testing.T) {
	virtualServer := NewVirtualServer(stge)

	cancel, addr := virtualServer.Serve("localhost:0")
	defer cancel()

	virtualServer.AddHandles("set", func(rw *RespWriter, req [][]byte) error {
		virtualServer.stge.Set(req[1], req[2])
		return rw.FlushString(Ok)
	})

	t.Run("cli", func(t *testing.T) {
		cli := NewClient("localhost:"+localAddr(addr),
			DialMaxIdelConns(1),
			DialReadTimeout(1*time.Second),
		)
		defer cli.Close()

		t.Run("set", func(t *testing.T) {
			if reply, err := String(cli.Do("set", "x", "123")); err != nil {
				t.Fatal(err)
			} else if reply != Ok {
				t.Fatalf("Expected value does not match: %s", reply)
			}
		})
	})

}

func TestClientToCluster(t *testing.T) {
	virtualServer := NewVirtualServer(stge)

	cancel, addr := virtualServer.Serve("localhost:0")
	defer cancel()

	virtualServer.AddHandles("cluster", func(rw *RespWriter, req [][]byte) error {
		b, err := virtualServer.stge.Cluster()
		if err != nil {
			return err
		}
		return rw.FlushBulk(b)
	})

	cli := NewClient("localhost:"+localAddr(addr), DialMaxIdelConns(1))
	defer cli.Close()

	t.Run("cluster", func(t *testing.T) {
		if reply, err := String(cli.Do("cluster", "nodes")); err != nil {
			t.Fatal(err)
		} else if reply != string(otherClusterNodeInfo) {
			t.Fatalf("Expected value does not match: %s", reply)
		}
	})
}

func TestClientToSentinel(t *testing.T) {
	virtualServer := NewVirtualServer(stge)

	cancel, addr := virtualServer.Serve("localhost:0")
	defer cancel()

	virtualServer.AddHandles("sentinel", func(rw *RespWriter, req [][]byte) error {
		b, err := virtualServer.stge.Sentinel()
		if err != nil {
			return err
		}
		return rw.FlushArray(b)
	})

	cli := NewClient("localhost:"+localAddr(addr), DialMaxIdelConns(1))
	defer cli.Close()

	t.Run("sentinel", func(t *testing.T) {
		if reply, err := cli.Do("sentinel", "master", "mymaster"); err != nil {
			t.Fatal(err)
		} else if v, ok := reply.([]interface{}); !ok {
			// _ = v
			t.Fatalf("Expected value does not match: %s", reply)
		} else {
			_ = v
		}
	})
}
