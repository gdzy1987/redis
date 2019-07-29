package topology

import (
	"encoding/json"
	"io"
)

type (
	Mode       int
	StateType  int
	StateEvent struct {
		StateType
		Error error
		Name  string
		Data  string
	}
)

const (
	SingleMode Mode = iota
	SentinelMode
	ClusterMode

	MasterStr = `master`
	SlaveStr  = `slave`
)

const (
	// node runing
	RUNING StateType = iota
	// During sent stop state
	STOPPED
	// Call node.collect timeout with 1 second
	TIMEOUT
	// Unexpected error occurred
	BROKEN
)

type (
	// service stop handle
	Stop func() error
	// Basic Run
	Basic interface {
		Run() Stop
	}
)

type ToplogyMapped map[*NodeInfo][]*NodeInfo

func (t *ToplogyMapped) Compares(cur *ToplogyMapped) (
	newNodes []*NodeInfo,
	oldNodes []*NodeInfo,
	hasChanged bool,
) {
	for i, v := range *t {
		if _, ok := (*cur)[i]; !ok {
			if oldNodes == nil {
				oldNodes = make([]*NodeInfo, 0)
			}
			oldNodes = append(oldNodes, i)
			hasChanged = true
		}
		_ = v
	}
	for i, v := range *cur {
		if _, ok := (*t)[i]; !ok {
			if newNodes == nil {
				newNodes = make([]*NodeInfo, 0)
			}
			newNodes = append(newNodes, i)
			hasChanged = true
		}
		_ = v
	}
	return
}

type Topologist interface {
	// Get redis real server topology
	Topology() *ToplogyMapped
	// Increment offset
	Increment(*NodeInfo, int64)
	// Group get current node groupAddr
	Group(*NodeInfo) *NodeInfoGroup
	// Offset get the string of offset
	Offset(*NodeInfo) string
	// marshal
	MarshalToWriter(io.Writer) error
	// The implementor needs to implement Basic interface template
	// Return the service callback method
	Basic
}

func NewTopologyist(mode Mode, pass string, addrs ...string) (t Topologist, err error) {
	switch mode {
	case ClusterMode:
		cAddrs, err := clusterAddr(pass, addrs...)
		if err != nil {
			return nil, err
		}
		t = CreateRedisCluster(pass, cAddrs)
	case SentinelMode:
		t = CreateRedisSentinel(pass, addrs...)
	case SingleMode:
		t = CreateRedisSingle(pass, addrs...)
	}
	return t, nil
}

func UnmarshalFromBytes(mode Mode, p []byte) (Topologist, error) {
	var th Topologist

	switch mode {
	case SingleMode:
		th = &RedisSingle{}
	case ClusterMode:
		th = &RedisCluster{}
	case SentinelMode:
		th = &RedisSentinel{}
	}
	err := json.Unmarshal(p, th)
	if err != nil {
		return nil, err
	}
	return th, nil
}
