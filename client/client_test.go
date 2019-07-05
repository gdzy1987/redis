package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
)

var (
	ErrCommand      = `(error) ERR wrong number of arguments for '%s' command`
	Nil             = `(nil)`
	Ok              = `OK`
	ClusterNodeInfo = `cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected`
)

var storage map[string]string = make(map[string]string)

func asyncServe() (func(), string) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err.Error())
		return func() { panic(err) }, ""
	}

	stopServer := func() {
		if err := listener.Close(); err != nil {
			panic(err)
		}
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v", err.Error())
				continue
			}
			go serverHandler(conn)
		}
	}()

	return stopServer, fmt.Sprintf("%s", listener.Addr())
}

func serverHandler(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	br := bufio.NewReader(conn)
	rReader := NewRespReader(br)

	bw := bufio.NewWriter(conn)
	rWriter := NewRespWriter(bw)

	req, err := rReader.ParseRequest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	fmt.Fprintf(os.Stdout, "parse request %s\n", req)

	handerRequest(rWriter, req)

}

func handerRequest(rWriter *RespWriter, req [][]byte) error {
	var (
		cmd  string
		args [][]byte
		err  error
	)

	if len(req) == 0 {
		cmd = ""
		args = nil
	} else {
		cmd = strings.ToLower(string(req[0]))
		args = req[1:]
	}

	key, value := "", ""
	switch cmd {
	case "ping":
		return rWriter.FlushString("PONG")
	case "get":
		if len(args) != 1 {
			return rWriter.FlushString(fmt.Sprintf(ErrCommand, "get"))
		}
		key = string(args[0])
		_, exist := storage[key]
		if !exist {
			return rWriter.FlushString(Nil)
		}
		err = rWriter.FlushString(storage[key])
	case "set":
		if len(args) < 2 {
			return rWriter.FlushString(fmt.Sprintf(ErrCommand, "set"))
		}
		key, value = string(args[0]), string(args[1])
		storage[key] = value
		return rWriter.FlushString(Ok)

	case "cluster":
		return rWriter.FlushString(ClusterNodeInfo)

	default:
		err = rWriter.FlushString("not support command")
	}
	return err
}

func localAddr(s string) string {
	if strings.Contains(s, "[::]") {
		// [::]:12345
		ss := strings.Split(s, ":")
		return ss[3]
	}
	return strings.Split(s, ":")[1]
}

func TestClientPing(t *testing.T) {
	stopServer, addr := asyncServe()
	defer stopServer()
	addr = localAddr(addr)

	cli := NewClient("localhost:" + addr)
	if reply, err := String(cli.Do("PING")); err != nil {
		t.Fatal(err)
	} else if reply != "PONG" {
		t.Fatal("Expected value does not match")
	}
	cli.Close()
}

func TestClientSet(t *testing.T) {
	stopServer, addr := asyncServe()
	defer stopServer()
	addr = localAddr(addr)
	cli := NewClient("localhost:" + addr)

	t.Run("set", func(t *testing.T) {
		if reply, err := String(cli.Do("set", "x", "123")); err != nil {
			t.Fatal(err)
		} else if reply != Ok {
			t.Fatalf("Expected value does not match: %s", reply)
		}
	})
	cli.Close()
}

func TestClientToCluster(t *testing.T) {
	stopServer, addr := asyncServe()
	defer stopServer()
	addr = localAddr(addr)
	cli := NewClient("localhost:" + addr)

	t.Run("cluster", func(t *testing.T) {
		if reply, err := String(cli.Do("cluster", "nodes")); err != nil {
			t.Fatal(err)
		} else if reply != ClusterNodeInfo {
			t.Fatalf("Expected value does not match: %s", reply)
		}
	})
	cli.Close()
}
