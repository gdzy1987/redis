package topology

type RedisCluster struct {
	// Multiple sets of msater slave information in one cluster
	Cluster map[string]*NodeInfoGroup `json:"Cluster"`
	// handler
	Handler ToplogyHandler

	stateC chan StateType

	stopped chan struct{}
}

func CreateRedisCluster(pass string, addrss [][]string) *RedisCluster {
	cluster := make(map[string]*NodeInfoGroup)
	for i := range addrss {
		addrGroup := addrss[i]
		key := concat(addrGroup...)
		cluster[key] = CreateMSNodeGroup(pass, addrGroup...)
	}
	return &RedisCluster{
		Cluster: cluster,
		stateC:  make(chan StateType),
		stopped: make(chan struct{}),
	}
}

func (s *RedisCluster) SetHandler(thr ToplogyHandler) { s.Handler = thr }

func (s *RedisCluster) Run() Stop {
	stop := func() error {
		for y := range s.Cluster {
			r := s.Cluster[y]
			for i := range r.Members {
				r.Members[i].Stop()
			}
		}
		// async stop
		go func() {
			s.stopped <- struct{}{}
			close(s.stopped)
		}()
		return nil
	}
	return stop
}

func (s *RedisCluster) monitor() error {
	for _, ngs := range s.Cluster {
		info, err := ProbeTopology(
			ngs.ArgumentPasswrod,
			ClusterMode,
			ngs.ArgumentAddrs...,
		)
		if err != nil {
			return err
		}
	}
}

func (s *RedisCluster) MasterNodeInfo() []*NodeInfo {
	nodes := make([]*NodeInfo, 0)

	for _, v := range s.Cluster {
		nodeGroup := v
		nodes = append(nodes, nodeGroup.Master())
	}

	return nodes
}

func (s *RedisCluster) SlaveNodeGroupInfo() []*NodeInfo {
	nodes := make([]*NodeInfo, 0)

	for _, v := range s.Cluster {
		nodeGroup := v
		nodes = append(nodes, nodeGroup.Slaves()...)
	}

	return nodes
}

func (s *RedisCluster) Notify() <-chan StateEvent {
	notifyC := make(chan StateEvent)

	go func() {
		for name, ngs := range s.Cluster {
			members := ngs.Members
			_, _ = members, name
		}
	}()

	return notifyC
}
