package replication

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	cli "github.com/dengzitong/redis/client"
	"github.com/dengzitong/redis/topology"
)

var OKReply = "OK"

type respServer struct {
	runID  string
	offset string

	co *cli.Conn

	bpool sync.Pool
}

func (resp *respServer) pull() error {
	err := resp.co.Send("psync", resp.runID, resp.offset)
	if err != nil {
		return err
	}
	return resp.co.DumpAndParse(resp.parse)
}

func (resp *respServer) close() {
	resp.co.Close()
}

func (resp *respServer) parse(rd io.Reader) error {
	p := resp.bpool.Get().([]byte)
	defer func() {
		resp.bpool.Put(p)
	}()
	_, err := io.ReadFull(rd, p)
	if err != nil {
		return err
	}

	fmt.Printf("%s", p)
	return nil
}

type Replication struct {
	toplogyist topology.Topologist

	tmapped *topology.ToplogyMapped

	changesC chan struct {
		ni []*topology.NodeInfo
		oi []*topology.NodeInfo
	}

	stge io.Writer
}

func NewReplication(t topology.Topologist, stge io.Writer) *Replication {
	repl := &Replication{
		changesC: make(chan struct {
			ni []*topology.NodeInfo
			oi []*topology.NodeInfo
		},
		),
		toplogyist: t,

		stge: stge,
	}

	return repl
}

func (r *Replication) prepareNode(master *topology.NodeInfo) (*respServer, error) {
	conn, ip, port, err := master.ExclusiveConn()
	if err != nil {
		return nil, err
	}
	if master.Ver > "4.0.0" {
		if ok, err := cli.String(conn.Do("replconf", "listening-port", port)); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf listening-port error")
		}

		if ok, err := cli.String(conn.Do("replconf", "ip-address", ip)); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf ip-address error")
		}

		if ok, err := cli.String(conn.Do("replconf", "capa", "eof")); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf capa eof error")
		}

		if ok, err := cli.String(conn.Do("replconf", "capa", "psync2")); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf capa psync2 error")
		}
	}
	runID := "?"
	var offset int64 = -1
	globalOffset := r.toplogyist.Group(master).GroupOffset
	if globalOffset > 0 {
		runID = master.Id
		offset = globalOffset
	}

	respSrv := &respServer{
		runID:  runID,
		offset: fmt.Sprintf("%d", offset),
		co:     conn,
		bpool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 1)
			},
		},
	}

	return respSrv, nil
}

func (r *Replication) Start() (func(), error) {
	stop := r.toplogyist.Run()

	cancel := func() {
		// finally,need to close the resource and persist the data.
		stop()
		err := r.toplogyist.MarshalToWriter(r.stge)
		if err != nil {
			println(err)
		}
	}

	tops := r.toplogyist.Topology()

	// first initialization
	for m, _ := range *tops {
		respSrv, err := r.prepareNode(m)
		if err != nil {
			return cancel, err
		}
		if err := respSrv.pull(); err != nil {
			return cancel, err
		}
	}
	r.tmapped = tops

	// 1 second check
	go r.everyOneSecCheck()

	go r.whenChange()

	return cancel, nil
}

func (r *Replication) whenChange() {
	for {
		notify, ok := <-r.changesC
		if !ok {
			return
		}

		for i := range notify.oi {
			node := notify.oi[i]
			r.toplogyist.Group(node).CloseAllMember()
		}

		for i := range notify.ni {
			node := notify.ni[i]
			respSrv, err := r.prepareNode(node)
			if err != nil {
				panic(err)
			}
			if err := respSrv.pull(); err != nil {
				panic(err)
			}
		}
	}
}

func (r *Replication) everyOneSecCheck() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			cur := r.toplogyist.Topology()
			ni, oi, hasChanged := r.tmapped.Compares(cur)
			if !hasChanged {
				continue
			}
			r.changesC <- struct {
				ni []*topology.NodeInfo
				oi []*topology.NodeInfo
			}{ni, oi}
		}
	}()
}
