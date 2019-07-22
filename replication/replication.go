package replication

import (
	"errors"
	"time"

	"github.com/dengzitong/redis/client"
	"github.com/dengzitong/redis/topology"
)

var OKReply = "OK"

type notify struct {
	ni []*topology.NodeInfo
	oi []*topology.NodeInfo
}

type Replication struct {
	Topologist topology.Topologist

	curTopMapped *topology.ToplogyMapped

	changesC chan notify

	respServer struct {
		cli *client.Client
		br  ByteReader
	}
}

func (r *Replication) prepareNode(master *topology.NodeInfo) error {
	repl_cli, ip, port, err := master.Client()
	if err != nil {
		return err
	}
	if master.Ver > "4.0.0" {
		if ok, err := client.String(repl_cli.Do("replconf", "listening-port", port)); err != nil {
			return err
		} else if ok != OKReply {
			return errors.New("replconf listening-port error")
		}

		if ok, err := client.String(repl_cli.Do("replconf", "ip-address", ip)); err != nil {
			return err
		} else if ok != OKReply {
			return errors.New("replconf ip-address error")
		}

		if ok, err := client.String(repl_cli.Do("replconf", "capa", "eof")); err != nil {
			return err
		} else if ok != OKReply {
			return errors.New("replconf capa eof error")
		}

		if ok, err := client.String(repl_cli.Do("replconf", "capa", "psync2")); err != nil {
			return err
		} else if ok != OKReply {
			return errors.New("replconf capa psync2 error")
		}
	}
	runID := ""
	if r.Topologist.Group(master).GroupOffset > 0 {
		runID = "?"
		repl_cli.Do("psync", runID, "-1")
	} else {
		repl_cli.Do("psync", master.Id, r.Topologist.Group(master).GroupOffset)
	}
	return nil
}

func (r *Replication) PrepareDump() error {
	stop := r.Topologist.Run()
	defer stop()
	tops := r.Topologist.Topology()

	// first initialization
	for m, _ := range *tops {
		err := r.prepareNode(m)
		if err != nil {
			return err
		}
	}

	r.curTopMapped = tops
	// curTopMapped

	return nil
}

func (r *Replication) onCheck() {
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
			r.changesC <- notify{
				ni, oi,
			}
		}
	}()
}
