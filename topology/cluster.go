package topology

import "net"

type RedisClusterTop struct {
	TopologyGroup map[string]*Topology `json:"topology_group"`
	Addrs         []string             `json:"addrs"`
}

func createRedisClusterTop() *RedisClusterTop {
	return &RedisClusterTop{}
}
func (s *RedisClusterTop) IsActivity() bool             { return false }
func (s *RedisClusterTop) OnConns() ([]net.Conn, error) { return nil, nil }
