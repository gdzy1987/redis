package topology

// redis single or master->slave architectural model
type RedisSingle struct {
	*NodeInfos `json:"nodes"`
}

func CreateRedisSingle() *RedisSingle {
	return &RedisSingle{}
}
