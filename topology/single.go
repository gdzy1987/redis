package topology

// redis single or master->slave architectural model
type RedisSingle struct {
	*NodeInfoGroup `json:"node_info_group"`
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

func (r *RedisSingle) MasterNodeInfo() []*NodeInfo {
	return []*NodeInfo{r.Master()}
}

func (r *RedisSingle) SlaveNodeGroupInfo() []*NodeInfo {
	return r.Slaves()
}
