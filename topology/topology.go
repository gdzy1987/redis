package topology

import (
	"errors"
	"strings"
	"sync/atomic"
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

func fpComparison(s, t string) bool { return strings.Contains(t, s) }

type NodeInfo struct {
	Version string `json:"node_version"`
	RunID   string `json:"node_runid"`
	IpAddr  string `json:"node_addr"`
}

type Topology struct {
	Mode        `json:"mode"`
	Fingerprint string      `json:"top_fingerprint"`
	Master      *NodeInfo   `json:"top_master"`
	Slaves      []*NodeInfo `json:"top_slaves"`
	Offset      int64       `json:"top_offset"`
}

// replace Master node
func (t *Topology) ResetMaster(n *NodeInfo) {
	t.Fingerprint = n.RunID
	t.Master = n
}

// cluster mode data node
func (t *Topology) CollectSlaves() {
	if t.Slaves == nil {
		t.Slaves = make([]*NodeInfo, 0)
	}
	// if there is old information you need to clear
	t.Slaves = t.Slaves[:0]

	// this is a serious mistake
	ms, err := ProbeNode(t.Master.IpAddr, "")
	if err != nil {
		panic(err)
	}
	replicationInfoMap, exits := ms[Replication]
	if !exits {
		panic("probe node info not exists replication selection")
	}
	slaveMap, err := ParsedReplicationInfo(replicationInfoMap)
	if err != nil {
		panic(err)
	}
	nodeInfos, err := ParsedSlaveInfo(slaveMap, "")
	if err != nil {
		panic(err)
	}
	for _, n := range nodeInfos {
		t.fingerprintCorrection(n.RunID)
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

func (t *Topology) UpdateOffset(i int64) {
	atomic.AddInt64(&t.Offset, i)
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
