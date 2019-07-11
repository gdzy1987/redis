package topology

import (
	"errors"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dengzitong/redis/client"
)

var (
	ErrProbe      = errors.New("probe error")
	ErrTopChanged = errors.New("top changed")
)

type Mode int

const (
	SingleMode Mode = iota
	SentinelMode
	ClusterMode

	MasterStr = `master`
	SlaveStr  = `slave`
)

func fpComparison(s, t string) bool { return strings.Contains(s, t) }

type NodeInfo struct {
	Id     string `json:"node_id"`
	Addr   string `json:"node_addr"`
	Pass   string `json:"node_password"`
	Ver    string `json:"node_version"`
	Offset int64  `json:"node_offset"`

	changed chan *NodeInfo
	stopped chan *NodeInfo

	c *client.Client
}

func CreateNodeInfo(addr string, pass string) (node *NodeInfo, stopped func()) {
	node = &NodeInfo{
		Addr: addr, Pass: pass,
		stopped: make(chan *NodeInfo),
	}

	node.prepare()

	node.surveySelf()
	stopped = func() { node.stop() }
	return node, stopped
}

func (n *NodeInfo) stop() { n.stopped <- n }

func (n *NodeInfo) surveySelf() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-n.stopped:
				n.c.Close()
				return
			default:
			}
			// ignore error
			reply, _ := client.String(n.c.Do("info"))
			if len(reply) < 1 {
				continue
			}
			mm := ParseNodeInfo(reply)
			server, exist := mm[Server]
			if !exist {
				continue
			}
			if len(n.Ver) < 1 {
				redis_version, exist := server["redis_version"]
				if !exist {
					continue
				}
				n.Ver = redis_version
			}
			if len(n.Id) < 1 {
				run_id, exist := server["run_id"]
				if !exist {
					continue
				}
				n.Id = run_id
			}

			replication, exist := mm[Replication]
			if !exist {
				continue
			}
			offset, exist := replication["master_repl_offset"]
			if !exist {
				continue
			}
			_offset, _ := strconv.ParseInt(offset, 10, 64)
			n.Offset = _offset
			<-ticker.C
		}
	}()
}

func (n *NodeInfo) prepare() {
	dialops := []client.DialOption{
		client.DialMaxIdelConns(1),
	}
	if len(n.Pass) > 0 {
		dialops = append(dialops,
			client.DialPassword(n.Pass),
		)
	}
	n.c = client.NewClient(n.Addr, dialops...)
}

type NodeInfos struct {
	UUID         string      `json:"uuid"`
	Members      []*NodeInfo `json:"nodes"`
	Num          int         `json:"num"`
	GlobalOffset int64       `json:"global_offset"`

	changeds []chan *NodeInfo
}

func NewNodeInfos() *NodeInfos {
	return &NodeInfos{
		UUID:     "",
		Members:  make([]*NodeInfo, 0),
		Num:      0,
		changeds: make([]chan *NodeInfo, 0),
	}
}

func (ns *NodeInfos) AddNodeInfo(ni *NodeInfo) {

}

type Topology struct {
	Mode        `json:"mode"`
	Fingerprint string      `json:"top_fingerprint"`
	Master      *NodeInfo   `json:"top_master"`
	Slaves      []*NodeInfo `json:"top_slaves"`
	Offset      int64       `json:"top_offset"`
	Password    string      `josn:"top_password"`

	cancel func()
}

// cluster mode data node
func (t *Topology) collect() {
	if t.Slaves == nil {
		t.Slaves = make([]*NodeInfo, 0)
	}
	// if there is old information you need to clear
	t.Slaves = t.Slaves[:0]

	// this is a serious mistake
	ms, err := ProbeNode(t.Master.Addr, t.Password)
	if err != nil {
		panic(err)
	}

	serverInfoMap, exist := ms[Server]
	if !exist {
		panic("probe node info server selection not exist")
	}

	_, exist = serverInfoMap["redis_version"]
	if !exist {
		panic("probe node info server.redis_version not exist")
	}

	t.Master.Ver = serverInfoMap["redis_version"]

	replicationInfoMap, exits := ms[Replication]
	if !exits {
		panic("probe node info replication selection not exist ")
	}

	slaveMap := ParseReplicationInfo(replicationInfoMap)
	if slaveMap == nil {
		return
	}

	nodeInfos, err := ParseSlaveInfo(slaveMap, "")
	if err != nil {
		panic(err)
	}

	for _, n := range nodeInfos {
		t.fingerprintCorrection(n.Id)
	}

	t.Slaves = nodeInfos
}

func (t *Topology) fingerprintCorrection(s string) {
	if strings.Contains(t.Fingerprint, s) {
		return
	}
	t.Fingerprint = strings.Join(
		[]string{t.Fingerprint, s},
		"-",
	)
}

func (t *Topology) UpdateOffset(i int64) { atomic.AddInt64(&t.Offset, i) }

func (t *Topology) alwaysCollect() {
	cancelSignal := make(chan struct{})
	go func() {
		tk := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-tk.C:
				t.collect()
			case <-cancelSignal:
				return
			}
		}
	}()
	t.cancel = func() { cancelSignal <- struct{}{} }
}

type (
	// service stop handle
	Stop func() error
	// Basic
	Basic interface {
		Run() Stop
	}
)

type TopologyHandler interface {
	// Topology service is ready served
	// When used, it needs to return true if the service is fully operational, otherwise it will block
	// When the topology of the backend changes
	// You need to notify the caller to re-adjust the entire cluster connection
	// Receive the topology information from the current architectural
	// And provide the current real master nodeInfo
	ReceiveNodeInfos() <-chan []*NodeInfo

	// The implementor needs to implement Basic interface template
	// Return the service callback method
	Basic
}

func NewTopologyHandler(mode Mode, addrs ...string) (t TopologyHandler, err error) {
	switch mode {
	case ClusterMode:
	case SentinelMode:
	case SingleMode:
	}
	return nil, nil
}
