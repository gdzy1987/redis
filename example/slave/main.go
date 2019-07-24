package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

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

	err = pool.Send("psync", "?", "-1")
	if err != nil {
		return err
	}

	f := func(rd io.Reader) error {
		p := make([]byte, 1, 1)
		_, err := io.ReadFull(rd, p)
		if err != nil {
			return err
		}

		fmt.Printf("%s", p)
		return nil
	}

	return pool.DumpAndParse(f)
}

func main() {
	_slave := &slave{}
	log.SetOutput(os.Stdout)
	err := _slave.listenAndSlaveSrve("127.0.0.1:6379")
	if err != nil {
		panic(err)
	}
}
