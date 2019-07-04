package topology

import (
	"encoding/json"
	"errors"
	"sync/atomic"
	"unsafe"
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

type NodeInfo struct {
	Version string `json:"node_version"`
	RunID   string `json:"node_runid"`
	IpAddr  string `json:"node_addr"`
}

type Topology struct {
	Mode   `json:"mode"`
	Master *NodeInfo   `json:"top_master"`
	Slaves []*NodeInfo `json:"top_slaves"`
	offset int64
}

type temporary struct {
	Mode   `json:"mode"`
	Master *NodeInfo   `json:"top_master"`
	Slaves []*NodeInfo `json:"top_slaves"`
	Offset int64       `json:"top_offset"`
}

func CreateTopFromStream(b []byte) (*Topology, error) {
	tt := &temporary{}
	if err := json.Unmarshal(b, tt); err != nil {
		return nil, err
	}
	return &Topology{tt.Mode, tt.Master, tt.Slaves, tt.Offset}, nil
}

func (t *Topology) UpdateOffset(i int64) {
	atomic.AddInt64(&t.offset, i)
}

func (t *Topology) String() string {
	tt := &temporary{
		t.Mode,
		t.Master,
		t.Slaves,
		t.offset,
	}
	bs, _ := json.Marshal(tt)
	return *(*string)(unsafe.Pointer(&bs))
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
	// Get the topology information from the current architectural
	// And provide the current real master nodeInfo
	GetNodeInfos() []*NodeInfo

	// Topology service is ready served
	// When used, it needs to return true if the service is fully operational, otherwise it will block
	IsActivity() <-chan struct{}

	// When the topology of the backend changes
	// You need to notify the caller to re-adjust the entire cluster connection
	IsChanged() <-chan struct{}

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

func testToplogyer() {
	top, err := NewTopologyHandler(ClusterMode, "10.1.1.228:21001,10.1.1.228:21002,10.1.1.228:21003")
	if err != nil {
		panic(err)
	}

	go func() {
		select {
		case <-top.IsActivity():
			// conns := top.OnConns()

		case <-top.IsChanged():
		}
	}()

	stop := top.Run()
	defer stop()

	select {}
}
