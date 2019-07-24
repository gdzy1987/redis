package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"strings"

	"github.com/dengzitong/redis/client"
)

type slave struct {
}

func (s *slave) listenAndSlaveSrve(masterAddr string) error {
	cli := client.NewClient(masterAddr, client.DialMaxIdelConns(1))

	pool, err := cli.Get()
	if err != nil {
		return err
	}

	ip, port, err := pool.Addr()
	if err != nil {
		return err
	}

	cli = pool.Client

	if ok, err := client.String(cli.Do("replconf", "listening-port", port)); err != nil {
		return err
	} else if ok != "OK" {
		return errors.New("replconf listening-port error")
	}

	if ok, err := client.String(cli.Do("replconf", "ip-address", ip)); err != nil {
		return err
	} else if ok != "OK" {
		return errors.New("replconf ip-address error")
	}

	if ok, err := client.String(cli.Do("replconf", "capa", "eof")); err != nil {
		return err
	} else if ok != "OK" {
		return errors.New("replconf capa eof error")
	}

	if ok, err := client.String(cli.Do("replconf", "capa", "psync2")); err != nil {
		return err
	} else if ok != "OK" {
		return errors.New("replconf capa psync2 error")
	}

	reply, err := client.String(cli.Do("psync", "?", "-1"))
	if err != nil {
		return err
	}
	log.Printf("%s", reply)

	ln, err := net.Listen("tcp", strings.Join([]string{ip, port}, ":"))
	if err != nil {
		return err
	}

	log.Printf("slave listen on %s\n", strings.Join([]string{ip, port}, ":"))

	co, err := ln.Accept()
	if err != nil {
		return err
	}

	return s.handle(co)
}

func (s *slave) handle(co net.Conn) error {
	br := bufio.NewReader(co)
	rReader := client.NewRespReader(br)
	for {
		bbs, err := rReader.ParseRequest()
		if err != nil {
			return err
		}
		log.Printf("xxxxx %s\n", bbs)
	}
}

func main() {
	_slave := &slave{}
	log.SetOutput(os.Stdout)
	err := _slave.listenAndSlaveSrve("127.0.0.1:6379")
	if err != nil {
		panic(err)
	}
}
