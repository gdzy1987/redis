package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/iauto360x/redis/client"
)

// Simulate the slave server, download only rdb and aof
type slave struct {
}

func (s *slave) slaveOfMaster(masterAddr string) error {
	cli := client.NewClient(masterAddr,
		client.DialMaxIdelConns(1),
	)

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

	bufPool := sync.Pool{
		New: func() interface{} {
			return make([]byte, 1)
		},
	}
	f := func(rd io.Reader) error {
		p := bufPool.Get().([]byte)
		defer func() {
			bufPool.Put(p)
		}()
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
	err := _slave.slaveOfMaster("127.0.0.1:6379")
	if err != nil {
		panic(err)
	}
}
