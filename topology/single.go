package topology

import "time"

// redis single or master->slave architectural model
type RedisSingle struct {
	*NodeInfos `json:"nodes"`
	stopped    chan *RedisSingle
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
		make(chan *RedisSingle),
	}
}

func (r *RedisSingle) Run() Stop {
	stop := func() error {
		for i := range r.Members {
			r.Members[i].Stop()
		}
		r.stopped <- r
		return nil
	}
	return stop
}

func (r *RedisSingle) ReceiveNodeInfos() <-chan []*NodeInfo {
	res := make(chan []*NodeInfo)
	go func() {
		secondTicker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-r.stopped:
				return
			default:
			}
			node := r.Master()
			if node == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			res <- []*NodeInfo{
				node,
			}
			<-secondTicker.C
		}
	}()
	return res
}
