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

	mu sync.Mutex

	initialized bool

	changed chan struct{}
	brokend chan struct{}
	stopped chan *RedisClusterTop
}

func CreateRedisClusterTop(addrs ...string) *RedisClusterTop {
	r := &RedisClusterTop{}
	r.TopGroup = make(map[string]*Topology)
	r.Addrs = addrs
	r.changed = make(chan struct{})
	r.brokend = make(chan struct{})
	return r
}

func (s *RedisClusterTop) Run() Stop {

	s.repeatPeek()

	return func() error {
		s.stopped <- s
		return nil
	}
}

func (s *RedisClusterTop) peek(ctx context.Context) chan error {
	errC := make(chan error)
	go func() {
		select {
		case <-ctx.Done():
			errC <- nil
			return

		default:
			topInfos, err := ProbeTopology(ClusterMode, s.Addrs...)
			if err != nil {
				errC <- err
				return
			}

			update := func(s *RedisClusterTop, ss []string) {
				s.mu.Lock()
				defer s.mu.Unlock()
				if _, exist := s.TopGroup[ss[0]]; !exist {
					// If it has been initialized, the topology does not exist in the dictionary.
					// The topology has changed.
					if s.initialized {
						errC <- ErrTopChanged
						return
					}
					s.TopGroup[ss[0]] = &Topology{
						Mode: ClusterMode,
						Master: &NodeInfo{
							RunID:  ss[0],
							IpAddr: ss[1],
						},
						offset: -1,
					}
				}
				errC <- err
				return
			}

			for i, _ := range topInfos {
				info := topInfos[i]
				ss := strings.Split(info, ",")
				if len(ss) < 2 {
					errC <- ErrProbe
					return
				}
				update(s, ss)
			}
		}
	}()
	return errC
}

func (s *RedisClusterTop) repeatPeek() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bgctx := context.Background()
				ctx, cancel := context.WithTimeout(bgctx, time.Second*5)
				defer cancel()

				switch <-s.peek(ctx) {
				case ErrTopChanged:
					s.changed <- struct{}{}
				case ErrProbe:
					s.brokend <- struct{}{}
					return
				}

				if !s.initialized {
					s.initialized = true
				}
			case <-s.stopped:
				return
			}
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
