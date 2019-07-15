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

type Topologyer interface {
	// Get master infos
	// Topology service is ready served
	// When used, it needs to return true if the service is fully operational, otherwise it will block
	// When the topology of the backend changes
	// You need to notify the caller to re-adjust the entire cluster connection
	// Receive the topology information from the current architectural
	// And provide the current real master nodeInfo
	// When the cluster mode has multiple masters
	MasterNodeInfo() []*NodeInfo

	SlaveNodeGroupInfo() []*NodeInfo

	// Unmarshal
	UnmarshalToWriter(io.Writer)

	// The implementor needs to implement Basic interface template
	// Return the service callback method
	Basic
}

func NewTopologyer(mode Mode, addrs ...string) (t Topologyer, err error) {
	switch mode {
	case ClusterMode:

	case SentinelMode:

	case SingleMode:
	}
	return nil, nil
}

func UnmarshalFromBytes(mode Mode, p []byte) (Topologyer, error) {
	var th Topologyer
	switch mode {
	case SingleMode:

	case ClusterMode:

	case SentinelMode:
	}
	err := json.Unmarshal(p, th)
	if err != nil {
		return nil, err
	}
	return th, nil
}

type ToplogyHandler interface {
	Start() error
	Shutdown() error
}
