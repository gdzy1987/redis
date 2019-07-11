package topology

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"
)

type nodeGroup struct {
	l sync.Mutex
	m map[string]*NodeInfo
}

func newNodeGroup() *nodeGroup {
	return &nodeGroup{
		l: sync.Mutex{},
		m: make(map[string]*NodeInfo),
	}
}

func (g *nodeGroup) put(node *NodeInfo) {
	g.l.Lock()
	defer g.l.Unlock()
	g.m[node.Id] = node
}

func (g *nodeGroup) diff(ng *nodeGroup) (increased []*NodeInfo, decreasd []*NodeInfo, hasdiff bool) {
	g.l.Lock()
	defer g.l.Unlock()
	ng.l.Lock()
	defer ng.l.Unlock()
	for name, nodeInfo := range ng.m {
		if _, exist := g.m[name]; !exist {
			if increased == nil {
				increased = make([]*NodeInfo, 0)
			}
			increased = append(increased, nodeInfo)
			hasdiff = true
		}
	}

	for name, nodeInfo := range g.m {
		if _, exist := ng.m[name]; !exist {
			if decreasd == nil {
				decreasd = make([]*NodeInfo, 0)
			}
			decreasd = append(decreasd, nodeInfo)
			hasdiff = true
		}
	}

	return increased, decreasd, hasdiff
}

type RedisClusterTop struct {
	Addrs  []string             `json:"addrs"`
	Pass   string               `json:"pass"`
	TGroup map[string]*Topology `json:"group"`

	incrNodeInfos []*NodeInfo
	decrNodeInfos []*NodeInfo

	mu  sync.Mutex
	one sync.Once

	changed chan struct{}
	stopped chan *RedisClusterTop

	initialized bool
}

func (s *RedisClusterTop) Format() string {
	p, err := json.Marshal(s.TGroup)
	if err != nil {
		panic(err)
	}
	return string(p)
}

func CreateRedisClusterTopFromAddrs(pwd string, addrs ...string) *RedisClusterTop {
	r := &RedisClusterTop{
		Pass: pwd,
		one:  sync.Once{},
	}
	r.TGroup = make(map[string]*Topology)
	r.Addrs = addrs
	r.changed = make(chan struct{})

	return r
}

func CreateRedisClusterTopFromStream(p []byte) (*RedisClusterTop, error) {
	rs := &RedisClusterTop{}
	err := json.Unmarshal(p, rs)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

func (s *RedisClusterTop) Run() Stop {
	s.repeatPeek()

	stop := func() error {
		s.stopped <- s
		return nil
	}

	return stop
}

func (s *RedisClusterTop) updateIncr(incrs []*NodeInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.incrNodeInfos = incrs
	for index := range incrs {
		node := incrs[index]
		t, ok := s.query(node)
		if ok {
			t.cancel()
			t.Master = node
			t.collect()
			continue
		}
		srcT := s.TGroup[t.Fingerprint]
		srcT = t
		srcT.alwaysCollect()
	}
}

func (s *RedisClusterTop) updateDecr(decr []*NodeInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.decrNodeInfos = decr
	for index := range decr {
		node := decr[index]
		t, ok := s.query(node)
		if !ok {
			continue
		}
		delete(s.TGroup, t.Fingerprint)
	}
}

func (s *RedisClusterTop) query(node *NodeInfo) (*Topology, bool) {
	for ffp, topology := range s.TGroup {
		if fpComparison(ffp, node.Id) {
			return topology, true
		}
	}
	t := &Topology{
		Fingerprint: node.Id,
		Mode:        ClusterMode,
		Master:      node,
		Password:    s.Pass,
		Slaves:      make([]*NodeInfo, 0),
		Offset:      -1,
	}
	t.collect()
	return t, false
}

func (s *RedisClusterTop) peek(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		topInfos, err := ProbeTopology(s.Pass, ClusterMode, s.Addrs...)
		if err != nil {
			return err
		}

		log.Printf("topinfo: %#v\n", topInfos)

		nng := newNodeGroup()
		for i := range topInfos {
			info := topInfos[i]
			ss := strings.Split(info, ",")
			if len(ss) < 2 {
				return ErrProbe
			}
			nng.put(&NodeInfo{Id: ss[0], Addr: ss[1]})
		}

		update := func(s *RedisClusterTop, nng *nodeGroup) error {
			sndg := newNodeGroup()
			for _, t := range s.TGroup {
				sndg.put(t.Master)
			}
			incrs, decrs, hasdiff := sndg.diff(nng)
			if hasdiff {
				s.updateIncr(incrs)
				s.updateDecr(decrs)
				if s.initialized {
					return ErrTopChanged
				}
			}
			return nil
		}
		if err := update(s, nng); err != nil {
			return err
		}
	}
	return nil
}

func (s *RedisClusterTop) repeatPeek() {
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopped:
				return
			default:
			}

			bgctx := context.Background()
			ctx, cancel := context.WithTimeout(bgctx, time.Second*1)
			defer cancel()
			switch s.peek(ctx) {
			case ErrTopChanged:
				s.changed <- struct{}{}
			case ErrProbe:
				panic(ErrProbe)
			}
			if !s.initialized {
				s.initialized = true
				s.one.Do(func() { s.changed <- struct{}{} })
			}
			<-ticker.C
		}
	}()
}

func (s *RedisClusterTop) ReceiveNodeInfos() (add <-chan []*NodeInfo, del <-chan []*NodeInfo) {
	additionNodes := make(chan []*NodeInfo)
	deleteNodes := make(chan []*NodeInfo)

	go func(s *RedisClusterTop) {
		for range s.changed {
			s.mu.Lock()
			addnodes := cuttingNodeSlice(s.incrNodeInfos)
			delnodes := cuttingNodeSlice(s.decrNodeInfos)
			s.mu.Unlock()
			additionNodes <- addnodes
			deleteNodes <- delnodes
		}
	}(s)

	return additionNodes, deleteNodes
}

func cuttingNodeSlice(s []*NodeInfo) []*NodeInfo {
	res := make([]*NodeInfo, len(s), len(s))
	for i := 0; i < len(s); i++ {
		res[i] = s[i]
	}
	s = s[:0]
	return res
}
