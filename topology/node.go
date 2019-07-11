package topology

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/dengzitong/redis/client"
)

var (
	ErrProbe      = errors.New("probe error")
	ErrTopChanged = errors.New("top changed")
)

func fpComparison(s, t string) bool { return strings.Contains(s, t) }

type NodeInfo struct {
	Id     string `json:"node_id"`
	Addr   string `json:"node_addr"`
	Pass   string `json:"node_password"`
	Ver    string `json:"node_version"`
	Offset int64  `json:"node_offset"`

	IsMaster bool `json:"is_master"`

	changed chan *NodeInfo

	c *client.Client
}

func CreateNodeInfo(addr string, pass string) *NodeInfo {
	return &NodeInfo{
		Addr: addr,
		Pass: pass,
	}
}

func (n *NodeInfo) Stop() { n.c.Close() }

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

func (n *NodeInfo) collect() {
	reply, _ := client.String(n.c.Do("info"))
	if len(reply) < 1 {
		return
	}
	mm := ParseNodeInfo(reply)
	server, exist := mm[Server]
	if !exist {
		return
	}
	if len(n.Ver) < 1 {
		redis_version, exist := server["redis_version"]
		if !exist {
			return
		}
		n.Ver = redis_version
	}
	if len(n.Id) < 1 {
		run_id, exist := server["run_id"]
		if !exist {
			return
		}
		n.Id = run_id
	}
	replication, exist := mm[Replication]
	if !exist {
		return
	}
	offset, exist := replication["master_repl_offset"]
	if !exist {
		return
	}
	_offset, _ := strconv.ParseInt(offset, 10, 64)
	n.Offset = _offset

	ismaster, exist := replication["role"]
	if !exist {
		return
	}
	n.IsMaster = (ismaster == MasterStr)
}

type NodeInfos struct {
	UUID    string      `json:"uuid"`
	Members []*NodeInfo `json:"nodes"` // index:0 is master node
	Offset  int64       `json:"offset"`

	MasterId string

	changeds []chan *NodeInfo
}

func CreateNodeInfos() *NodeInfos {
	return &NodeInfos{
		UUID:     NewSUID().String(),
		Members:  make([]*NodeInfo, 0),
		changeds: make([]chan *NodeInfo, 0),
	}
}

func (ns *NodeInfos) Put(n *NodeInfo) {
	ns.Members = append(ns.Members, n)
}

func (ns *NodeInfos) Len() int {
	return len(ns.Members)
}

func (ns *NodeInfos) Less(i, j int) bool {
	return ns.Members[i].IsMaster
}

func (ns *NodeInfos) Swap(i, j int) {
	ns.Members[i], ns.Members[j] = ns.Members[j], ns.Members[i]
}

func (ns *NodeInfos) Master() *NodeInfo {
	if !ns.hasMaster() {
		for _, member := range ns.Members {
			member.collect()
		}
	}
	sort.Sort(ns)
	return ns.Members[0]
}

func (ns *NodeInfos) hasMaster() bool {
	for _, member := range ns.Members {
		if member.IsMaster {
			return true
		}
	}
	return false
}
