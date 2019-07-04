package topology

import (
	"context"
	"strings"
	"sync"
	"time"
)

type RedisClusterTop struct {
	TopGroup map[string]*Topology `json:"top_group"`
	Addrs    []string             `json:"addrs"`

	mu          sync.Mutex
	initialized bool
	changed     chan struct{}
	brokend     chan struct{}
}

func createRedisClusterTop(addrs ...string) *RedisClusterTop {
	r := &RedisClusterTop{}
	r.TopGroup = make(map[string]*Topology)
	r.Addrs = addrs
	r.changed = make(chan struct{})
	r.brokend = make(chan struct{})
	return r
}

func (s *RedisClusterTop) peekAndCompare() {
	update := func(ctx context.Context, s *RedisClusterTop) error {
		select {
		case <-ctx.Done():
			return nil
		default:
			topInfos, err := ProbeTopology(ClusterMode, s.Addrs...)
			if err != nil {
				return err
			}
			for i, _ := range topInfos {
				info := topInfos[i]
				ss := strings.Split(info, ",")
				if len(ss) < 2 {
					return ErrProbe
				}
				isChange := func() error {
					s.mu.Lock()
					defer s.mu.Unlock()
					if _, exist := s.TopGroup[ss[0]]; !exist {
						if s.initialized {
							return ErrTopChanged
						}
						masterNodeInfo := &NodeInfo{
							RunID:  ss[0],
							IpAddr: ss[1],
						}
						s.TopGroup[ss[0]] = &Topology{
							Mode:   ClusterMode,
							Master: masterNodeInfo,
							offset: -1,
						}
					}
					return nil
				}
				if err = isChange(); err != nil {
					return err
				}
			}
		}
		return nil
	}

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			bgctx := context.Background()
			ctx, cancel := context.WithTimeout(bgctx, time.Second*5)
			defer cancel()

			if err := update(ctx, s); err != nil {
				switch err {
				case ErrTopChanged:
					s.changed <- struct{}{}
				case ErrProbe:
					s.brokend <- struct{}{}
					return
				}
			}

			if !s.initialized {
				s.initialized = true
			}
			<-ticker.C
		}
	}()
}

func (s *RedisClusterTop) GetNodeInfos() []*NodeInfo {
	s.mu.Lock()
	length := len(s.Addrs) / 2
	nodes := make([]*NodeInfo, length, length)
	i := 0
	for _, top := range s.TopGroup {
		nodes[i] = top.Master
		i++
	}
	s.mu.Unlock()
	return nodes
}

func (s *RedisClusterTop) IsActivity() <-chan struct{} { return s.brokend }

func (s *RedisClusterTop) IsChanged() <-chan struct{} { return s.changed }
