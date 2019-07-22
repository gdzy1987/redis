package topology

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

type RedisSentinel struct {
	*NodeInfoGroup `json:"sentinel"`
}

func CreateRedisSentinel(pass string, addrs ...string) *RedisSentinel {
	infoSlice, err := probeTopology(pass, SentinelMode, addrs...)
	if err != nil {
		panic(err)
	}
	if _, ok := infoSlice.([][]string); !ok {
		panic("probe sentinel topology error")
	}
	nodeInfos := infoSlice.([][]string)
	realAddrs := make([]string, len(nodeInfos), len(nodeInfos))
	for i := range nodeInfos {
		realAddrs[i] =
			strings.Replace(
				fieldSplicing(sliceStr2Dict(nodeInfos[i]), "ip", "port"), ",", ":", 1)
	}

	_MSNodeGroup := CreateMSNodeGroup(pass, realAddrs...)
	return &RedisSentinel{_MSNodeGroup}
}

func (s *RedisSentinel) Run() Stop {
	stop := func() error {
		for i := range s.NodeInfoGroup.Members {
			s.NodeInfoGroup.Members[i].Stop()
		}
		return nil
	}
	return stop
}

func (s *RedisSentinel) Topology() *ToplogyMapped {
	master := s.NodeInfoGroup.Master()
	slaves := s.NodeInfoGroup.Slaves()
	return &ToplogyMapped{
		master: slaves,
	}
}

func (s *RedisSentinel) MarshalToWriter(dst io.Writer) error {
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

func (s *RedisSentinel) Increment(n *NodeInfo, offset int64) {
	s.NodeInfoGroup.Update(offset)
}

func (s *RedisSentinel) Offset(n *NodeInfo) string {
	return s.NodeInfoGroup.Offset()
}

func (s *RedisSentinel) Group(n *NodeInfo) *NodeInfoGroup {
	return s.NodeInfoGroup
}
