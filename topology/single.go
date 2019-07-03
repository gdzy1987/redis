package topology

import "net"

type RedisSingleTop struct {
	TopologyGroup []*Topology `json:"topology_group"`
	Addrs         []string    `json:"addrs"`
}

func createRedisSingleTop() *RedisSingleTop {
	return &RedisSingleTop{}
}

func (s *RedisSingleTop) IsActivity() bool             { return false }
func (s *RedisSingleTop) OnConns() ([]net.Conn, error) { return nil, nil }
