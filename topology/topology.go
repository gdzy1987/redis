package topology

import "net"

type (
	Mode     int
	RoleType int
)

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

type RedisCluster struct {
	TopologyGroup map[string]*Topology `json:"topology_group"`
	Addrs         []string             `json:"addrs"`
}

type RedisSentinel struct {
	TopologyGroup []*Topology `json:"topology_group"`
	Addrs         []string    `json:"addrs"`
}

type RedisSingle struct {
	TopologyGroup []*Topology `json:"topology_group"`
	Addrs         []string    `json:"addrs"`
}

type RedisTopology interface {
	IsActivity() bool
	OnConnect() (net.Conn, error)
}

func CreateTopology(mode Mode, addrs ...string) (RedisTopology, error) {

	return nil, nil
}
