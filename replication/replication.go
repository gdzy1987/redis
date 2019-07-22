package replication

import (
	"errors"
	"time"

	"github.com/dengzitong/redis/client"
	"github.com/dengzitong/redis/topology"
)

var OKReply = "OK"

type respServer struct {
	cli *client.Client
	br  ByteReader
}

func (resp *respServer) start() error {
	return nil
}

type Replication struct {
	Topologist topology.Topologist

	curTopMapped *topology.ToplogyMapped

	changesC chan struct {
		ni []*topology.NodeInfo
		oi []*topology.NodeInfo
	}

	respSrv *respServer
}

func NewReplication(topologist topology.Topologist) *Replication {
	repl := &Replication{
		changesC: make(chan struct {
			ni []*topology.NodeInfo
			oi []*topology.NodeInfo
		},
		),
		Topologist: topologist,
	}

	return repl
}

func (r *Replication) prepareNode(master *topology.NodeInfo) (*respServer, error) {
	repl_cli, ip, port, err := master.Client()
	if err != nil {
		return nil, err
	}
	if master.Ver > "4.0.0" {
		if ok, err := client.String(repl_cli.Do("replconf", "listening-port", port)); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf listening-port error")
		}

		if ok, err := client.String(repl_cli.Do("replconf", "ip-address", ip)); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf ip-address error")
		}

		if ok, err := client.String(repl_cli.Do("replconf", "capa", "eof")); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf capa eof error")
		}

		if ok, err := client.String(repl_cli.Do("replconf", "capa", "psync2")); err != nil {
			return nil, err
		} else if ok != OKReply {
			return nil, errors.New("replconf capa psync2 error")
		}
	}
	runID := ""
	if r.Topologist.Group(master).GroupOffset > 0 {
		runID = "?"
		repl_cli.Do("psync", runID, "-1")
	} else {
		repl_cli.Do("psync", master.Id, r.Topologist.Group(master).GroupOffset)
	}
	return &respServer{cli: repl_cli}, nil
}

func (r *Replication) DumpAndParse() {
	stop := r.Topologist.Run()
	defer stop()
	tops := r.Topologist.Topology()

	// first initialization
	for m, _ := range *tops {
		respSrv, err := r.prepareNode(m)
		if err != nil {
			panic(err)
		}
		if err := respSrv.start(); err != nil {
			panic(err)
		}
	}
	r.curTopMapped = tops

	// 1 second check
	r.tickerOneSecCheck()

	for {
		<-r.changesC

	}
}

func (r *Replication) tickerOneSecCheck() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			cur := r.Topologist.Topology()
			ni, oi, hasChanged := r.curTopMapped.Compares(cur)
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
