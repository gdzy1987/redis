package replication

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	cli "github.com/dengzitong/redis/client"
	tp "github.com/dengzitong/redis/topology"
)

var OKReply = "OK"

type Outer interface {
	Receive(Command)
}

type Replication struct {
	toplogyist tp.Topologist

	tmapped *tp.ToplogyMapped

	changesC chan struct {
		ni []*tp.NodeInfo
		oi []*tp.NodeInfo
	}

	stge io.Writer
	out  Outer
}

func NewReplication(t tp.Topologist, stge io.Writer, out Outer) *Replication {
	repl := &Replication{
		toplogyist: t,

		changesC: make(chan struct {
			ni []*tp.NodeInfo
			oi []*tp.NodeInfo
		},
		),
		stge: stge,
		out:  out,
	}

	return repl
}

func (r *Replication) prepareNode(master *tp.NodeInfo) (*respServer, error) {
	co, ip, port, err := master.ExclusiveConn()
	if err != nil {
		return nil, err
	}
	if master.Ver > "4.0.0" {
		if ok, err := cli.String(co.Do("replconf", "listening-port", port)); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf listening-port error")
		}

		if ok, err := cli.String(co.Do("replconf", "ip-address", ip)); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf ip-address error")
		}

		if ok, err := cli.String(co.Do("replconf", "capa", "eof")); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf capa eof error")
		}

		if ok, err := cli.String(co.Do("replconf", "capa", "psync2")); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf capa psync2 error")
		}
	}
	runID := "?"
	var offset int64 = -1
	tgroup := r.toplogyist.Group(master)
	if tgroup.GroupOffset > 0 {
		runID = master.Id
		offset = tgroup.GroupOffset
	}

	respSrv := &respServer{
		runID:  runID,
		offset: fmt.Sprintf("%d", offset),
		co:     co,
		node:   master,
		repl:   r,
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
			group := r.toplogyist.Group(node)
			group.CloseAllMember()
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
				ni []*tp.NodeInfo
				oi []*tp.NodeInfo
			}{ni, oi}
		}
	}()
}
