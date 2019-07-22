package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type DataStorage interface {
	// data operate command
	Set([]byte, []byte) error
	Get([]byte) ([]byte, error)
	Hset([]byte, []byte, []byte) error
	Hget([]byte, []byte) ([]byte, error)

	// system command
	Cluster() ([]byte, error)
	Sentinel() ([]interface{}, error)
	Info() ([]byte, error)
}

type HandlesFunc = func(rWriter *RespWriter, req [][]byte) error

type VirtualServer struct {
	DataStorage

	mu      sync.Mutex
	addr    string
	handles map[string]HandlesFunc
}

func NewVirtualServer(stge DataStorage) *VirtualServer {
	return &VirtualServer{
		mu: sync.Mutex{},

		DataStorage: stge,
		handles:     make(map[string]HandlesFunc),
	}
}

func (v *VirtualServer) AddHandles(name string, h HandlesFunc) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.handles[name] = h
}

func (v *VirtualServer) Serve(addr string) (cancel func(), port string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return func() { panic(err) }, ""
	}

	cancel = func() {
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
			go v.serverHandler(conn)
		}
	}()

	return cancel, fmt.Sprintf("%s", listener.Addr())
}

func (v *VirtualServer) serverHandler(conn net.Conn) {
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
		return
	}

	v.handerRequest(rWriter, req)
}

func (v *VirtualServer) handerRequest(rWriter *RespWriter, req [][]byte) error {
	var cmd string

	if len(req) == 0 {
		cmd = ""
	} else {
		cmd = strings.ToLower(string(req[0]))
	}
	_, exist := v.handles[cmd]
	if !exist {
		return errors.New("undefine handles")
	}

	return v.handles[cmd](rWriter, req)
}
