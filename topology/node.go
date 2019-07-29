package topology

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/dengzitong/redis/client"
)

func concat(strs ...string) string { return strings.Join(strs, ":") }

type NodeInfo struct {
	Id     string `json:"node_id"`
	Addr   string `json:"node_addr"`
	Pass   string `json:"node_password"`
	Ver    string `json:"node_version"`
	Offset int64  `json:"node_offset"`

	IsMaster bool `json:"is_master"`

	c *client.Client
}

func CreateNodeInfo(addr string, pass string) *NodeInfo {
	return &NodeInfo{
		Addr:   addr,
		Pass:   pass,
		Offset: -1,
	}
}

func (n *NodeInfo) ExclusiveConn() (*client.Conn, string, string, error) {
	pc, err := n.c.Get()
	if err != nil {
		return nil, "", "", err
	}
	ip, port, err := pc.Conn.Addr()
	if err != nil {
		return nil, "", "", err
	}
	return pc.Conn, ip, port, nil
}

func (n *NodeInfo) Stop() {
	defer func() {
		// repeated closing error
		if err := recover(); err != nil {
			// println("try to close a client that has been closed")
		}
	}()
	n.c.Close()
}

func (n *NodeInfo) prepare() {
	dialops := []client.DialOption{
		client.DialMaxIdelConns(2),
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
	mm := parseNodeInfo(reply)
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

type NodeInfoGroup struct {
	UUID string `json:"uuid"`
	// Index:0 is master node
	Members []*NodeInfo `json:"members_node"`
	// Group offset
	// In the process of switching,
	// If the value is greater than the offset value of the instance to be run, then the value is discarded,
	// And the full amount of data of the new instance is retrieved directly from -1
	// Otherwise it continues from the current offset.
	GroupOffset int64 `json:"group_offset"`
	// [{ip:port:runid}.String(),...]
	MemberIds []string `json:"member_ids"`
	// Node instance count
	MemberCnt int32 `json:"member_cnt"`
	// Identify current master run_id
	MasterId string
	// External incoming argsments address
	ArgumentAddrs []string
	// External incoming argsments password
	ArgumentPasswrod string
}

func CreateNodeInfoGroup() *NodeInfoGroup {
	return &NodeInfoGroup{
		UUID:        NewSUID().String(),
		Members:     make([]*NodeInfo, 0),
		MemberIds:   make([]string, 0),
		GroupOffset: -1,
	}
}

func CreateMSNodeGroup(pass string, addrs ...string) *NodeInfoGroup {
	nig := CreateNodeInfoGroup()

	for _, addr := range addrs {
		node := CreateNodeInfo(addr, pass)
		node.prepare()
		node.collect()
		nig.Put(node)
	}
	nig.ArgumentAddrs = addrs
	nig.ArgumentPasswrod = pass

	return nig
}

func (ns *NodeInfoGroup) Put(n *NodeInfo) {
	atomic.AddInt32(&ns.MemberCnt, 1)
	ns.MemberIds = append(
		ns.MemberIds,
		concat(n.Addr, n.Id),
	)
	ns.Members = append(ns.Members, n)
}

func (ns *NodeInfoGroup) Len() int {
	return len(ns.Members)
}

func (ns *NodeInfoGroup) Less(i, j int) bool {
	return ns.Members[i].IsMaster
}

func (ns *NodeInfoGroup) Swap(i, j int) {
	ns.Members[i], ns.Members[j] = ns.Members[j], ns.Members[i]
}

func (ns *NodeInfoGroup) Master() *NodeInfo {
	if !ns.hasMaster() {
		for _, member := range ns.Members {
			member.collect()
		}
	}
	sort.Sort(ns)
	return ns.Members[0]
}

func (ns *NodeInfoGroup) hasMaster() bool {
	for _, member := range ns.Members {
		if member.IsMaster {
			return true
		}
	}
	return false
}

func (ns *NodeInfoGroup) Slaves() []*NodeInfo {
	if !ns.hasMaster() {
		for _, member := range ns.Members {
			member.collect()
		}
	}
	sort.Sort(ns)
	return ns.Members[1:]
}

func (ns *NodeInfoGroup) CloseAllMember() {
	for _, member := range ns.Members {
		member.Stop()
	}
}

func (ns *NodeInfoGroup) Update(offset int64) {
	ns.GroupOffset = offset
}

func (ns *NodeInfoGroup) Offset() string {
	return fmt.Sprintf("%d", ns.GroupOffset)
}
