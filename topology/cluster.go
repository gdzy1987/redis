package topology

import (
	"bytes"
	"encoding/json"
	"io"
)

var _ = Topologist(&RedisCluster{})

type RedisCluster struct {
	// Multiple sets of msater slave information in one cluster
	Cluster map[string]*NodeInfoGroup `json:"Cluster"`
}

func CreateRedisCluster(pass string, addrss [][]string) *RedisCluster {
	cluster := make(map[string]*NodeInfoGroup)

	for i := range addrss {
		groupAddr := addrss[i]
		mgs := CreateMSNodeGroup(pass, groupAddr...)
		key := createkeyList()
		for i := range mgs.Members {
			key.add(mgs.Members[i].Id)
		}
		keystr := key.String()
		cluster[keystr] = mgs
	}

	return &RedisCluster{Cluster: cluster}
}

func (s *RedisCluster) Run() Stop {
	stop := func() error {
		for y := range s.Cluster {
			r := s.Cluster[y]
			for i := range r.Members {
				r.Members[i].Stop()
			}
		}
		return nil
	}
	return stop
}

func (s *RedisCluster) masterNodeInfo() []*NodeInfo {
	nodes := make([]*NodeInfo, 0)

	for _, v := range s.Cluster {
		nodeGroup := v
		nodes = append(nodes, nodeGroup.Master())
	}

	return nodes
}

func (s *RedisCluster) slaveNodeGroupInfo(n *NodeInfo) []*NodeInfo {
	nodes := make([]*NodeInfo, 0)

	for i := range s.Cluster {
		members := s.Cluster[i].Members
		lkey := createkeyList()
		for _, ng := range members {
			lkey.add(ng.Id)
		}
		if !lkey.include(n.Id) {
			continue
		}
		nodes = members
		break
	}
	return nodes
}

func (s *RedisCluster) Topology() map[*NodeInfo][]*NodeInfo {
	res := make(map[*NodeInfo][]*NodeInfo)
	mns := s.masterNodeInfo()
	for i := range mns {
		m := mns[i]
		s := s.slaveNodeGroupInfo(m)
		res[m] = s
	}
	return res
}

// masrshal
func (s *RedisCluster) MarshalToWriter(dst io.Writer) error {
	p, err := json.Marshal(s)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, bytes.NewBuffer(p))
	if err != nil {
		return err
	}
	return nil
}
