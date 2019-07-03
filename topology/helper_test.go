package topology

import (
	"bytes"
	"net"
	"testing"
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
