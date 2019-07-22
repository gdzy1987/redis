package replication

import (
	"errors"

	"github.com/dengzitong/redis/client"
	"github.com/dengzitong/redis/topology"
)

var OKReply = "OK"

type Replication struct {
	Topologist topology.Topologist
	ByteReader
}

func (r *Replication) DumpAndParse() error {
	stop := r.Topologist.Run()
	defer stop()
	tops := r.Topologist.Topology()

	for master, slaves := range *tops {
		cli, err := master.Client()
		if err != nil {
			return err
		}
		pool, err := cli.Get()
		if err != nil {
			return err
		}
		ip, port, err := pool.Conn.Addr()
		if err != nil {
			return err
		}
		replication_client, err := client.UseAlreadyConn(pool.Conn)
		if err != nil {
			return err
		}

		if master.Ver > "4.0.0" {
			if ok, err := client.String(replication_client.Do("replconf", "listening-port", port)); err != nil {
				return err
			} else if ok != OKReply {
				return errors.New("replconf listening-port error")
			}

			if ok, err := client.String(replication_client.Do("replconf", "ip-address", ip)); err != nil {
				return err
			} else if ok != OKReply {
				return errors.New("replconf ip-address error")
			}

			if ok, err := client.String(replication_client.Do("replconf", "capa", "eof")); err != nil {
				return err
			} else if ok != OKReply {
				return errors.New("replconf capa eof error")
			}

			if ok, err := client.String(replication_client.Do("replconf", "capa", "psync2")); err != nil {
				return err
			} else if ok != OKReply {
				return errors.New("replconf capa psync2 error")
			}
		}
		runID := ""
		if r.Topologist.Group(master).GroupOffset > 0 {
			runID = "?"
			replication_client.Do("psync", runID, "-1")
		} else {
			replication_client.Do("psync", master.Id, r.Topologist.Group(master).GroupOffset)
		}
		_, _, _ = ip, port, slaves
	}

	return nil
}
