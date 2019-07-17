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
	// Basic
	Basic interface {
		Run() Stop
	}
)

type Topologist interface {
	// Topology service is ready served
	// When used, it needs to return true if the service is fully operational, otherwise it will block
	// When the topology of the backend changes
	// You need to notify the caller to re-adjust the entire cluster connection
	// Receive the topology information from the current architectural
	// And provide the current real master nodeInfo
	// When the cluster mode has multiple masters
	Topology() map[*NodeInfo][]*NodeInfo

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
