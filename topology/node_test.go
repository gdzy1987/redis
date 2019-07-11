package topology

import (
	"strings"
	"testing"

	"github.com/dengzitong/redis/client"
)

func port(s string) string {
	if strings.Contains(s, "[::]") {
		// [::]:12345
		ss := strings.Split(s, ":")
		return ss[3]
	}
	return strings.Split(s, ":")[1]
}

func TestNodeInfo(t *testing.T) {

	testinfo := `# Server
redis_version:4.0.2
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:ec5ba1d66550e200
redis_mode:cluster
os:Linux 4.4.0-116-generic x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.8.1
process_id:3022
run_id:c41a2899e5c155a02a38e26450a916d1466379eb
tcp_port:7001
uptime_in_seconds:354360
uptime_in_days:4
hz:10
lru_clock:2036497
executable:/opt/redis7001/src/redis-server
config_file:/opt/redis7001/redis7001.conf

# Replication
role:master
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=3758993694,lag=1
master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:3759000664
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:268435456
repl_backlog_first_byte_offset:3490565209
repl_backlog_histlen:268435456`

	virtualSrv := client.NewVirtualServer(nil)

	virtualSrv.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(testinfo))
	})
	cancel, addr := virtualSrv.Serve("localhost:0")
	defer cancel()

	node := CreateNodeInfo("localhost:"+port(addr), "")
	defer node.Stop()
	node.prepare()
	node.collect()

	if node.IsMaster == false {
		t.Fatal("expected master value is true")
	}
}

func TestNodeInfosSort(t *testing.T) {
	nis := &NodeInfos{
		Members: []*NodeInfo{
			&NodeInfo{
				Id:       "A",
				IsMaster: false,
			},
			&NodeInfo{
				Id:       "B",
				IsMaster: false,
			},
			&NodeInfo{
				Id:       "C",
				IsMaster: true,
			},
		},
	}

	m := nis.Master()
	if m.Id != "C" {
		t.Fatal("sort not implement")
	}

}

var (
	node1 = `# Server
redis_version:4.0.2
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:ec5ba1d66550e200
redis_mode:cluster
os:Linux 4.4.0-116-generic x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.8.1
process_id:3022
run_id:c41a2899e5c155a02a38e26450a916d1466379ea
tcp_port:7001
uptime_in_seconds:354360
uptime_in_days:4
hz:10
lru_clock:2036497
executable:/opt/redis7001/src/redis-server
config_file:/opt/redis7001/redis7001.conf

# Replication
role:master
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=3758993694,lag=1
master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:3759000664
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:268435456
repl_backlog_first_byte_offset:3490565209
repl_backlog_histlen:268435456`

	node2 = `# Server
redis_version:4.0.2
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:ec5ba1d66550e200
redis_mode:cluster
os:Linux 4.4.0-116-generic x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.8.1
process_id:3022
run_id:c41a2899e5c155a02a38e26450a916d1466379eb
tcp_port:7001
uptime_in_seconds:354360
uptime_in_days:4
hz:10
lru_clock:2036497
executable:/opt/redis7001/src/redis-server
config_file:/opt/redis7001/redis7001.conf

# Replication
role:slave
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=3758993694,lag=1
master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:3759000664
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:268435456
repl_backlog_first_byte_offset:3490565209
repl_backlog_histlen:268435456`

	node3 = `# Server
redis_version:4.0.2
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:ec5ba1d66550e200
redis_mode:cluster
os:Linux 4.4.0-116-generic x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.8.1
process_id:3022
run_id:c41a2899e5c155a02a38e26450a916d1466379ec
tcp_port:7001
uptime_in_seconds:354360
uptime_in_days:4
hz:10
lru_clock:2036497
executable:/opt/redis7001/src/redis-server
config_file:/opt/redis7001/redis7001.conf

# Replication
role:slave
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=3758993694,lag=1
master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:3759000664
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:268435456
repl_backlog_first_byte_offset:3490565209
repl_backlog_histlen:268435456`

	node6 = `# Server
redis_version:4.0.2
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:ec5ba1d66550e200
redis_mode:cluster
os:Linux 4.4.0-116-generic x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.8.1
process_id:3022
run_id:c41a2899e5c155a02a38e26450a916d1466379ed
tcp_port:7001
uptime_in_seconds:354360
uptime_in_days:4
hz:10
lru_clock:2036497
executable:/opt/redis7001/src/redis-server
config_file:/opt/redis7001/redis7001.conf

# Replication
role:master
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=3758993694,lag=1
master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:3759000664
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:268435456
repl_backlog_first_byte_offset:3490565209
repl_backlog_histlen:268435456`

	node4 = `# Server
redis_version:4.0.2
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:ec5ba1d66550e200
redis_mode:cluster
os:Linux 4.4.0-116-generic x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.8.1
process_id:3022
run_id:c41a2899e5c155a02a38e26450a916d1466379ee
tcp_port:7001
uptime_in_seconds:354360
uptime_in_days:4
hz:10
lru_clock:2036497
executable:/opt/redis7001/src/redis-server
config_file:/opt/redis7001/redis7001.conf

# Replication
role:slave
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=3758993694,lag=1
master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:3759000664
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:268435456
repl_backlog_first_byte_offset:3490565209
repl_backlog_histlen:268435456`

	node5 = `# Server
redis_version:4.0.2
redis_git_sha1:00000000
redis_git_dirty:0
redis_build_id:ec5ba1d66550e200
redis_mode:cluster
os:Linux 4.4.0-116-generic x86_64
arch_bits:64
multiplexing_api:epoll
atomicvar_api:atomic-builtin
gcc_version:4.8.1
process_id:3022
run_id:c41a2899e5c155a02a38e26450a916d1466379ef
tcp_port:7001
uptime_in_seconds:354360
uptime_in_days:4
hz:10
lru_clock:2036497
executable:/opt/redis7001/src/redis-server
config_file:/opt/redis7001/redis7001.conf

# Replication
role:slave
connected_slaves:1
slave0:ip=10.1.1.228,port=7004,state=online,offset=3758993694,lag=1
master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
master_replid2:0000000000000000000000000000000000000000
master_repl_offset:3759000664
second_repl_offset:-1
repl_backlog_active:1
repl_backlog_size:268435456
repl_backlog_first_byte_offset:3490565209
repl_backlog_histlen:268435456`
)

func threeNode() ([]string, func()) {
	s1 := client.NewVirtualServer(nil)
	s1.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node1))
	})
	cancel1, s1addr := s1.Serve("localhost:0")

	s2 := client.NewVirtualServer(nil)
	s2.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node2))
	})
	cancel2, s2addr := s2.Serve("localhost:0")

	s3 := client.NewVirtualServer(nil)
	s3.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node3))
	})
	cancel3, s3addr := s3.Serve("localhost:0")

	return []string{s1addr, s2addr, s3addr}, func() {
		defer func() {
			for _, cancel := range []func(){cancel1, cancel2, cancel3} {
				cancel()
			}
		}()
	}
}

func sixNode() ([][]string, func()) {
	s1 := client.NewVirtualServer(nil)
	s1.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node1))
	})
	cancel1, s1addr := s1.Serve("localhost:0")

	s2 := client.NewVirtualServer(nil)
	s2.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node2))
	})
	cancel2, s2addr := s2.Serve("localhost:0")

	s3 := client.NewVirtualServer(nil)
	s3.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node3))
	})
	cancel3, s3addr := s3.Serve("localhost:0")

	s4 := client.NewVirtualServer(nil)
	s4.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node4))
	})
	cancel4, s4addr := s4.Serve("localhost:0")

	s5 := client.NewVirtualServer(nil)
	s5.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node5))
	})
	cancel5, s5addr := s5.Serve("localhost:0")

	s6 := client.NewVirtualServer(nil)
	s6.AddHandles("info", func(rw *client.RespWriter, req [][]byte) error {
		return rw.FlushBulk([]byte(node6))
	})
	cancel6, s6addr := s6.Serve("localhost:0")

	return [][]string{[]string{s1addr, s2addr, s3addr}, []string{s4addr, s5addr, s6addr}}, func() {
		defer func() {
			for _, cancel := range []func(){
				cancel1,
				cancel2,
				cancel3,
				cancel4,
				cancel5,
				cancel6,
			} {
				cancel()
			}
		}()
	}

}

func TestNodeInfos(t *testing.T) {

	addrs, cancels := threeNode()
	defer cancels()

	nodes := CreateNodeInfos()
	for _, addr := range addrs {
		node := CreateNodeInfo(addr, "")
		node.prepare()
		node.collect()
		nodes.Put(node)
	}

	masterNode := nodes.Master()

	if masterNode.Id != "c41a2899e5c155a02a38e26450a916d1466379ea" {
		t.Fatal("The expected master id is inconsistent with the actual")
	}
}
