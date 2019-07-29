package replication

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	cli "github.com/iauto360x/redis/client"
	"github.com/iauto360x/redis/topology"
)

type respServer struct {
	co *cli.Conn

	runID  string
	offset string

	repl  *Replication
	bpool sync.Pool
	node  *topology.NodeInfo

	ismark bool
	Nop
}

func (resp *respServer) pull() error {
	err := resp.co.Send("psync", resp.runID, resp.offset)
	if err != nil {
		return err
	}
	resp.loopAck()
	return resp.co.DumpAndParse(resp.parse)
}

func (resp *respServer) loopAck() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			if resp.ismark {
				offset := resp.repl.toplogyist.Offset(resp.node)
				err := resp.co.Send("replconf", "ack", offset)
				if err != nil {
					panic(err)
				}
			}
			<-ticker.C
		}
	}()
}

func (resp *respServer) close() {
	resp.co.Close()
}

func (resp *respServer) incr(n int64) {
	resp.repl.toplogyist.Increment(resp.node, n)
}

func (resp *respServer) parse(rd io.Reader) error {
	redis := NewReader(rd)
	v, n, err := redis.ReadValue()
	if err != nil {
		return err
	}
	if resp.ismark {
		resp.incr(int64(n))
	}
	_type := v.Type()
	switch _type {
	case SimpleString:
		if strings.Contains(v.String(), "CONTINUE") {
			resp.ismark = true
		}
		return nil

	case BulkString:
		str := v.String()
		// selecting db will occur in versions below 5.0
		if strings.Contains(str, "SELECT") {
			// skip master DB
			DB, _, err := redis.ReadValue()
			if err != nil {
				return err
			}
			dbnum := DB.String()
			// *2\r\n$6\r\nSELECT\r\n${n}\r\n{num}\r\n
			cmd := NewCommand(Select, []byte(dbnum))
			resp.incr(int64(10 + len(dbnum)))
			resp.repl.out.Receive(cmd)

			// need to convert the ping command to aof byte accumulating offset
		} else if strings.Contains(str, "PING") {
			resp.incr(int64(4)) // *1\r\n$4\r\nPING\r\n length = 14
		}
		return nil

	case Array:
		ss := strings.Split(v.String(), " ")
		_type := CommandTypeMap[strings.ToLower(ss[0])]
		if _type == Ping {
			return nil
		}
		bbs := make([][]byte, len(ss)-1, len(ss)-1)
		for i := 1; i < len(ss); i++ {
			bbs[i-1] = []byte(ss[i])
		}
		resp.repl.out.Receive(Command{
			_type,
			bbs,
		})
		return nil

	case Rdb:
		_, offset := v.ReplInfo()
		resp.incr(offset)

		err = decodeStream(rd, resp)
		if err != nil {
			return err
		}
		_, _, err := redis.ReadValue()
		if err != nil {
			return err
		}

		resp.ismark = true
		return nil
	}

	return errors.New("unknow command")
}

func (r *respServer) BeginDatabase(n int) {
	r.repl.out.Receive(NewCommand(Select, []byte(fmt.Sprintf("%d", n))))
}

func (r *respServer) Set(key, value []byte, expiry int64) {
	r.repl.out.Receive(NewCommand(Set, key, value))
}

func (r *respServer) Hset(key, field, value []byte) {
	r.repl.out.Receive(NewCommand(Hset, key, value))
}

func (r *respServer) Sadd(key, member []byte) {
	r.repl.out.Receive(NewCommand(Sadd, key, member))
}

func (r *respServer) Rpush(key, value []byte) {
	r.repl.out.Receive(NewCommand(Rpush, key, value))
}

func float64ToByte(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)

	return bytes
}

func (r *respServer) Zadd(key []byte, score float64, member []byte) {
	r.repl.out.Receive(NewCommand(Zadd, float64ToByte(score), member))
}

func (r *respServer) EndDatabase(n int) {
	if n != 0 {
		r.repl.out.Receive(NewCommand(Select, []byte(fmt.Sprintf("%d", 0))))
	}
}
