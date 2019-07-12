package topology

import "time"

// redis single or master->slave architectural model
type RedisSingle struct {
	*NodeInfos `json:"nodes"`
}

func CreateRedisSingle(pass string, addrs ...string) *RedisSingle {
	nodeInfos := CreateNodeInfos()
	for _, addr := range addrs {
		node := CreateNodeInfo(addr, pass)
		node.prepare()
		nodeInfos.Put(node)
	}
	return &RedisSingle{
		nodeInfos,
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

func (r *RedisSingle) ReceiveNodeInfos() <-chan []*NodeInfo {
	res := make(chan []*NodeInfo)
	go func() {
	WAIT:
		node := r.Master()
		if node == nil {
			time.Sleep(1 * time.Second)
			goto WAIT
		}
		res <- []*NodeInfo{
			node,
		}
	}()
	return res
}
