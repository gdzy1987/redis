package topology

type RedisSentinel struct {
	*RedisSingle
}

func CreateRedisSentinel(pass string, addrs ...string) *RedisSentinel {
	single := CreateRedisSingle(pass, addrs...)
	return &RedisSentinel{single}
}
