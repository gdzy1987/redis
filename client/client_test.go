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
	ErrCommand = `(error) ERR wrong number of arguments for '%s' command`
	Nil        = `(nil)`
	Ok         = `OK`
)

var storage map[string]string = make(map[string]string)

func asyncServe() (func(), string) {
	listener, err := net.Listen("tcp", ":0")
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

	cli := NewClient(":" + addr)
	if reply, err := String(cli.Do("PING")); err != nil {
		t.Fatal(err)
	} else if reply != "PONG" {
		t.Fatal("Expected value does not match")
	}
}

func TestClientSet(t *testing.T) {
	stopServer, addr := asyncServe()
	defer stopServer()
	addr = localAddr(addr)

	cli := NewClient(":6379")
	defer cli.Close()

	t.Run("set", func(t *testing.T) {
		if reply, err := String(cli.Do("set", "x", "123")); err != nil {
			t.Fatal(err)
		} else if reply != Ok {
			t.Fatalf("Expected value does not match: %s", reply)
		}
	})
}
