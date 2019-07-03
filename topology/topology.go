package topology

import "net"

type (
	Mode     int
	RoleType int
)

type RClient interface {
	Do(cmd string, args ...interface{}) (interface{}, error)
}

const (
	SingleMode Mode = iota
	SentinelMode
	ClusterMode

	Master RoleType = iota
	Slave

	MasterStr = `master`
	SlaveStr  = `slave`
)

type NodeInfo struct {
	Version string   `json:"node_version"`
	RunID   string   `json:"node_runid"`
	IpAddr  string   `json:"node_addr"`
	Role    RoleType `json:"node_role"`
}

func NewNodeInfo(addr string) (*NodeInfo, error) {
	nodeInfo := &NodeInfo{
		IpAddr: addr,
	}
	return nodeInfo, nil
}

type Topology struct {
	Mode       `json:"mode"`
	Name       string      `json:"top_name"`
	Master     *NodeInfo   `json:"top_master"`
	SlaveGroup []*NodeInfo `json:"slave_group"`
}

func NewTopology(addr string) (*Topology, error) {
	topology := new(Topology)
	return topology, nil
}

type (
	// service stop handle
	Stop func() error
	// on outer function send a signal to called
	Cancel func() chan struct{}
	Basic  interface {
		Run() (Stop, Cancel)
	}
)

type TopologyHandler interface {
	// Get the connection from the current topology
	// And provide the current real master connection by the topology service
	OnConns() ([]net.Conn, error)

	// Topology server is runing
	IsActivity() bool

	// When the topology of the backend changes
	// You need to notify the caller to re-adjust the entire cluster connection
	IsChangeC() chan struct{}

	// The implementor needs to implement Basic interface template
	// Return the service callback method
	Basic
}

func CreateTopologyer(mode Mode, addrs ...string) (t TopologyHandler, err error) {
	switch mode {
	case ClusterMode:
	case SentinelMode:
	case SingleMode:
	}
	return nil, nil
}
