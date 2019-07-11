package topology

type RedisCluster struct {
	Cluster []*RedisSingle `json:"Cluster"`
}

func CreateRedisCluster(pass string, addrss [][]string) *RedisCluster {
	clusterSingle := make([]*RedisSingle, len(addrss), len(addrss))
	for index := range addrss {
		clusterSingle[index] = CreateRedisSingle(pass, addrss[index]...)
	}
	return &RedisCluster{clusterSingle}
}

func (s *RedisCluster) Run() Stop {
	stop := func() error {
		for y := range s.Cluster {
			r := s.Cluster[y]
			for i := range r.Members {
				r.Members[i].Stop()
			}
			r.stopped <- r
		}
		return nil
	}
	return stop
}

func (s *RedisCluster) ReceiveNodeInfos() <-chan []*NodeInfo {
	res := make(chan []*NodeInfo)
	go func() {
		nodes := []*NodeInfo{}
		for index := range s.Cluster {
			nodes = append(nodes, <-s.Cluster[index].ReceiveNodeInfos()...)
		}
		res <- nodes
	}()

	return res
}
