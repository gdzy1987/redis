package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var storage map[string]string = make(map[string]string)

type handlesFunc = func(rWriter *RespWriter, req [][]byte) error

type VirtualServer struct {
	mu      sync.Mutex
	handles map[string]handlesFunc
}

func NewVirtualServer() *VirtualServer {
	return &VirtualServer{
		mu:      sync.Mutex{},
		handles: make(map[string]handlesFunc),
	}
}

func (v *VirtualServer) AddHandles(name string, h handlesFunc) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.handles[name] = h
}

func (v *VirtualServer) Serve() (cancel func(), port string) {
	listener, err := net.Listen("tcp", "localhost:0")
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
	}
	fmt.Fprintf(os.Stdout, "parse request %s\n", req)

	v.handerRequest(rWriter, req)

}

func (v *VirtualServer) handerRequest(rWriter *RespWriter, req [][]byte) error {
	var cmd string
	if len(req) == 0 {
		cmd = ""
	} else {
		cmd = strings.ToLower(string(req[0]))
	}
	return v.handles[cmd](rWriter, req)
}
