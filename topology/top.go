package topology

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"
)

type Mode int

const (
	SingleMode Mode = iota
	SentinelMode
	ClusterMode

	MasterStr = `master`
	SlaveStr  = `slave`
)

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

type Topologyer interface {
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

func NewTopologyer(mode Mode, addrs ...string) (t Topologyer, err error) {
	switch mode {
	case ClusterMode:

	case SentinelMode:

	case SingleMode:
	}
	return nil, nil
}

func UnmarshalFromBytes(mode Mode, p []byte) (Topologyer, error) {
	var th Topologyer
	switch mode {
	case SingleMode:

	case ClusterMode:

	case SentinelMode:
	}
	err := json.Unmarshal(p, th)
	if err != nil {
		return nil, err
	}
	return th, nil
}
