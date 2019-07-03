package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	redis "github.com/dengzitong/redis/client"
)

var (
	ErrCommand = `(error) ERR wrong number of arguments for '%s' command`
	Nil        = `(nil)`
	Ok         = `OK`
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

func main() {
	stop := asyncServe("127.0.0.1:9999")
	defer stop()
	select {}
}
