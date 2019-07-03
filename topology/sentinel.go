package topology

import "net"

type RedisSentinelTop struct {
	TopologyGroup []*Topology `json:"topology_group"`
	Addrs         []string    `json:"addrs"`
}

func createRedisSentinelTop() *RedisSentinelTop {
	return &RedisSentinelTop{}
}
func (s *RedisSentinelTop) IsActivity() bool             { return false }
func (s *RedisSentinelTop) OnConns() ([]net.Conn, error) { return nil, nil }
