package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	redis "github.com/iauto360x/redis/client"
)

var (
	ErrCommand = "(error) ERR wrong number of arguments for '%s' command"
	Nil        = "(nil)"
	Ok         = "OK"

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

	otherClusterNodeInfo = `cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected`
)

var storage map[string]string = make(map[string]string)

func asyncServe(addr string) func() {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err.Error())
		return func() {
			panic(err)
		}
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

	return stopServer
}

func serverHandler(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	br := bufio.NewReader(conn)
	rReader := redis.NewRespReader(br)

	bw := bufio.NewWriter(conn)
	rWriter := redis.NewRespWriter(bw)

	req, err := rReader.ParseRequest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}
	handerRequest(rWriter, req)
}

func handerRequest(rWriter *redis.RespWriter, req [][]byte) error {
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
		err = rWriter.FlushBulk([]byte(storage[key]))
	case "set":
		if len(args) < 2 {
			return rWriter.FlushString(fmt.Sprintf(ErrCommand, "set"))
		}
		key, value = string(args[0]), string(args[1])
		storage[key] = value
		return rWriter.FlushString(Ok)
	case "cluster":
		return rWriter.FlushString(string(otherClusterNodeInfo))
	case "sentinel":
		return rWriter.FlushArray(SentinelNodeInfo)
	default:
		err = rWriter.FlushString("not support command")
	}
	return err
}

func main() {
	stop := asyncServe("127.0.0.1:9999")
	defer stop()
	select {}
}
