package topology

import (
	"bytes"
	"encoding/json"
	"io"
)

// redis single or master->slave architectural model
type RedisSingle struct {
	*NodeInfoGroup `json:"single"`
}

func CreateRedisSingle(pass string, addrs ...string) *RedisSingle {
	NodeInfoGroup := CreateNodeInfoGroup()
	for _, addr := range addrs {
		node := CreateNodeInfo(addr, pass)
		node.prepare()
		NodeInfoGroup.Put(node)
	}
	return &RedisSingle{
		NodeInfoGroup,
	}
}

func (r *RedisSingle) Run() Stop {
	stop := func() error {
		for i := range r.Members {
			r.Members[i].Stop()
		}
		return nil
	}
	return stop
}

func (r *RedisSingle) masterNodeInfo() *NodeInfo {
	return r.Master()
}

func (r *RedisSingle) slaveNodeGroupInfo() []*NodeInfo {
	return r.Slaves()
}

func (r *RedisSingle) Topology() map[*NodeInfo][]*NodeInfo {
	res := make(map[*NodeInfo][]*NodeInfo)
	key, value := r.masterNodeInfo(), r.slaveNodeGroupInfo()
	res[key] = value
	return res
}

func (r *RedisSingle) MarshalToWriter(dst io.Writer) error {
	p, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, bytes.NewBuffer(p))
	if err != nil {
		return err
	}
	return nil
}
