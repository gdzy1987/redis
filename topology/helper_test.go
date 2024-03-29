package topology

import (
	"bytes"
	"net"
	"testing"
	"time"
)

const info = `# Server
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

# Clients
connected_clients:214
client_longest_output_list:0
client_biggest_input_buf:0
blocked_clients:0

# Memory
used_memory:283793920
used_memory_human:270.65M
used_memory_rss:307539968
used_memory_rss_human:293.29M
used_memory_peak:296925040
used_memory_peak_human:283.17M
used_memory_peak_perc:95.58%
used_memory_overhead:274182194
used_memory_startup:1437888
used_memory_dataset:9611726
used_memory_dataset_perc:3.40%
total_system_memory:270373158912
total_system_memory_human:251.80G
used_memory_lua:41984
used_memory_lua_human:41.00K
maxmemory:0
maxmemory_human:0B
maxmemory_policy:volatile-lru
mem_fragmentation_ratio:1.08
mem_allocator:jemalloc-4.0.3
active_defrag_running:0
lazyfree_pending_objects:0

# Persistence
loading:0
rdb_changes_since_last_save:31622118
rdb_bgsave_in_progress:0
rdb_last_save_time:1561963230
rdb_last_bgsave_status:ok
rdb_last_bgsave_time_sec:0
rdb_current_bgsave_time_sec:-1
rdb_last_cow_size:25513984
aof_enabled:1
aof_rewrite_in_progress:0
aof_rewrite_scheduled:0
aof_last_rewrite_time_sec:1
aof_current_rewrite_time_sec:-1
aof_last_bgrewrite_status:ok
aof_last_write_status:ok
aof_last_cow_size:11870208
aof_current_size:131151347
aof_base_size:2325391
aof_pending_rewrite:0
aof_buffer_length:0
aof_rewrite_buffer_length:0
aof_pending_bio_fsync:0
aof_delayed_fsync:13

# Stats
total_connections_received:91995
total_commands_processed:34574580
instantaneous_ops_per_sec:60
total_net_input_bytes:2666927431
total_net_output_bytes:4060828581
instantaneous_input_kbps:4.40
instantaneous_output_kbps:7.09
rejected_connections:0
sync_full:1
sync_partial_ok:0
sync_partial_err:1
expired_keys:31411990
evicted_keys:0
keyspace_hits:1136
keyspace_misses:1963494
pubsub_channels:0
pubsub_patterns:0
latest_fork_usec:2821
migrate_cached_sockets:0
slave_expires_tracked_keys:0
active_defrag_hits:0
active_defrag_misses:0
active_defrag_key_hits:0
active_defrag_key_misses:0

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
repl_backlog_histlen:268435456

# CPU
used_cpu_sys:2877.08
used_cpu_user:1509.24
used_cpu_sys_children:0.76
used_cpu_user_children:0.80

# Cluster
cluster_enabled:1

# Keyspace
db0:keys=6246,expires=4451,avg_ttl=184203049`

func listener(addrs ...string) []net.Listener {
	listeners := make([]net.Listener, 0)
	for _, addr := range addrs {
		listen, err := net.Listen("tcp", addr)
		if err != nil {
			//
		}
		listeners = append(listeners, listen)
	}
	return listeners
}

func TestPing(t *testing.T) {
	addrs := []string{
		"127.0.0.1:50001",
		"127.0.0.1:50002",
		"127.0.0.1:50003",
		"127.0.0.1:2181",
		"127.0.0.1:2180",
	}

	lns := listener(addrs...)
	defer func() {
		for _, ln := range lns {
			ln.Close()
		}
	}()

	t.Run("ping", func(t *testing.T) {
		_addrs, err := ping(addrs...)
		if err != nil {
			t.Fatal(err)
		}
		if len(_addrs) != len(lns) {
			t.Fatal("expected addrs length not equal")
		}
	})
}

func TestCommandSize(t *testing.T) {
	b, size := convertAcmdSize("ping")
	if size != 14 {
		t.Fatal("expected length not equal.")
	}
	if !bytes.Equal(b, []byte("*1\r\n$4\r\nping\r\n")) {
		t.Fatal("expected value not equal.")
	}
}

var (
	otherClusterNodeInfo = `cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected`

	SentinelNodeInfo = []string{
		"name",
		"mymaster",
		"ip",
		"10.1.1.228",
		"port",
		"8003",
		"runid",
		"8be9401d095b9072b2e67c59fea68894e14da193",
		"flags",
		"master",
		"link-pending-commands",
		"0",
		"link-refcount",
		"1",
		"last-ping-sent",
		"0",
		"last-ok-ping-reply",
		"848",
		"last-ping-reply",
		"848",
		"down-after-milliseconds",
		"60000",
		"info-refresh",
		"10000",
		"role-reported",
		"master",
		"role-reported-time",
		"222861310",
		"config-epoch",
		"4",
		"num-slaves",
		"124",
		"num-other-sentinels",
		"2",
		"quorum",
		"2",
		"failover-timeout",
		"180000",
		"parallel-syncs",
		"1",
	}
)

func TestProbeTopParseInfoForMasters(t *testing.T) {

	clusterAddrs, err := parseInfoForMasters(ClusterMode, otherClusterNodeInfo)
	if err != nil {
		t.Fatal(err)
	} else if len(clusterAddrs) < 1 {
		t.Fatal("need at least 1 master node")
	}

	sentinelAddrs, err := parseInfoForMasters(SentinelMode, SentinelNodeInfo)
	if err != nil {
		t.Fatal(err)
	} else if len(sentinelAddrs) < 1 {
		t.Fatal("need at least 1 master node")
	}

	singleAddrs, err := parseInfoForMasters(SingleMode, "localhost:1234")
	if err != nil {
		t.Fatal(err)
	} else if len(singleAddrs) < 1 {
		t.Fatal("need at least 1 master node")
	}
}

func TestExec(t *testing.T) {
	err := execTimeout(time.Second*1, "ps", "-ef")
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseNodeInfo(t *testing.T) {

	m := parseNodeInfo(info)

	selections := []sectionType{
		Server,
		Clients,
		Memory,
		Persistence,
		Stats,
		Replication,
		CPU,
		Cluster,
		Keyspace,
	}
	for _, selection := range selections {
		if _, exits := m[selection]; !exits {
			t.Fatalf("expected '%s' selection not exist", m[Server])
		}
	}

	replicationInfo := m[Replication]
	t.Run("Parse replication info", func(t *testing.T) {
		replm := parseReplicationInfo(replicationInfo)
		if replm == nil {
			t.Fatal("empty map")
		}
		if _, exist := replm["slave0"]; !exist {
			t.Fatal("expected slave0 info not exist")
		}
	})
}

func TestListKey(t *testing.T) {
	kl := createKeyList().
		add("a").
		add("b").
		add("c")

	if kl.String() != "a_b_c" {
		t.Fatal("expected value not equal")
	}

	if !kl.include("a") && !kl.include("b") && !kl.include("c") {
		t.Fatal("include error")
	}
}

func TestParseCmdReplyToClusterNode(t *testing.T) {
	info := `cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected
`

	mp, err := parseCmdReplyToClusterNode(info)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range mp {
		if k.include("cebd9205cbde0d1ec4ad75600849a88f1f6294f6") {
			switch v[1].Id {
			case "a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab":
				break
			default:
				t.Fatal("slave error")
			}
			if k.String() != "cebd9205cbde0d1ec4ad75600849a88f1f6294f6_a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab" {
				t.Fatal("string() error")
			}
		} else if k.include("c6d165b72cfcd76d7662e559dc709e00e3dabf03") {
			switch v[1].Id {
			case "656042ad560b887164138a19dab2502154f8b039":
				break
			default:
				t.Fatal("slave error")
			}
			if k.String() != "c6d165b72cfcd76d7662e559dc709e00e3dabf03_656042ad560b887164138a19dab2502154f8b039" {
				t.Fatal("string() error")
			}
		} else if k.include("885493415bea22919fc9ce83836a9e6a8d0c1314") {
			switch v[1].Id {
			case "62bd020a2a5121a27c0e5540d1f0d4bba08cebb2":
				break
			default:
				t.Fatal("slave error")
			}
			if k.String() != "885493415bea22919fc9ce83836a9e6a8d0c1314_62bd020a2a5121a27c0e5540d1f0d4bba08cebb2" {
				t.Fatal("string() error")
			}
		} else {
			t.Fatal("master error")
		}

		_ = v
	}
}

// var (
// 	SentinelMaster = []interface{}{
// 		[]byte("name"),
// 		[]byte("mymaster"),
// 		[]byte("ip"),
// 		[]byte("127.0.0.1"),
// 		[]byte("port"),
// 		[]byte("8001"),
// 		[]byte("runid"),
// 		[]byte("8be9401d095b9072b2e67c59fea68894e14da193"),
// 		[]byte("flags"),
// 		[]byte("master"),
// 	}
// 	SentinelSlave1 = []interface{}{
// 		[]byte("name"),
// 		[]byte("mymaster"),
// 		[]byte("ip"),
// 		[]byte("127.0.0.1"),
// 		[]byte("port"),
// 		[]byte("8002"),
// 		[]byte("runid"),
// 		[]byte("8be9401d095b9072b2e67c59fea68894e14da194"),
// 		[]byte("flags"),
// 		[]byte("master"),
// 	}
// 	SentinelSlave2 = []interface{}{
// 		[]byte("name"),
// 		[]byte("mymaster"),
// 		[]byte("ip"),
// 		[]byte("127.0.0.1"),
// 		[]byte("port"),
// 		[]byte("8003"),
// 		[]byte("runid"),
// 		[]byte("8be9401d095b9072b2e67c59fea68894e14da195"),
// 		[]byte("flags"),
// 		[]byte("slave"),
// 	}
// )

// func TestProbeTopology(t *testing.T) {
// 	x, err := probeTopology("", SentinelMode, "10.1.1.228:21001")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	_ = x
// }
