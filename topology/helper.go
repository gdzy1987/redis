package topology

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	c "github.com/dengzitong/redis/client"
)

type sectionType string

const (
	Server      sectionType = "Server"
	Clients     sectionType = "Clients"
	Memory      sectionType = "Memory"
	Persistence sectionType = "Persistence"
	Stats       sectionType = "Stats"
	Replication sectionType = "Replication"
	CPU         sectionType = "CPU"
	Cluster     sectionType = "Cluster"
	Keyspace    sectionType = "Keyspace"
)

type keyList []string

func createkeyList() *keyList {
	kl := make(keyList, 0)
	return &kl
}

func (kl *keyList) add(s string) *keyList {
	*kl = append(*kl, s)
	return kl
}

func (kl *keyList) include(s string) bool {
	for i := range *kl {
		if s == (*kl)[i] {
			return true
		}
	}
	return false
}

func (kl *keyList) String() string {
	return strings.Join(*kl, "_")
}

//exec another process
//if wait d Duration, it will kill the process
//d is <= 0, wait forever
func ExecTimeout(d time.Duration, name string, args ...string) error {
	cmd := exec.Command(name, args...)

	if err := cmd.Start(); err != nil {
		return err
	}

	if d <= 0 {
		return cmd.Wait()
	}

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	duration := time.Second * d
	after := time.NewTimer(duration)
	select {
	case <-after.C:
		cmd.Process.Kill()

		//wait goroutine return
		return <-done
	case err := <-done:
		return err
	}
}

// Ping analog icmp detection
// Probe multiple addresses and return the address port that is currently responding within seconds
func Ping(addrs ...string) ([]string, error) {
	res := make([]string, 0)
	mu := sync.Mutex{}

	additionRes := func(addr string) {
		mu.Lock()
		defer mu.Unlock()
		res = append(res, addr)
	}

	diag := func(addr string) bool {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return false
		}
		defer conn.Close()
		return true
	}

	ping := func(ctx context.Context, addr string, f func(addr string)) {
		select {
		case <-ctx.Done():
			return
		default:
			if diag(addr) {
				f(addr)
			}
		}
	}

	wg := sync.WaitGroup{}
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			defer cancel()
			ping(ctx, addr, additionRes)
		}(addr)
	}
	wg.Wait()

	return res, nil
}

// Convert simple string command to Aof Command and calculate size
func ConvertAcmdSize(cmd string, args ...interface{}) ([]byte, int) {
	buf := bytes.NewBuffer(nil)

	buf.Write([]byte{'*'})
	lengthStr := fmt.Sprintf("%d", len(args)+1)
	buf.Write([]byte(lengthStr))
	buf.Write([]byte{'\r', '\n'})

	buklCmd := func(v interface{}, buf *bytes.Buffer) {
		var str []byte
		buf.Write([]byte("$"))
		switch v.(type) {
		case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
			str = []byte(fmt.Sprintf("%d", v))
		case string:
			str = []byte(v.(string))
		case []byte:
			str = v.([]byte)
		}
		buf.Write([]byte(fmt.Sprintf("%d", len(str))))
		buf.Write([]byte{'\r', '\n'})
		buf.Write([]byte(str))
		buf.Write([]byte{'\r', '\n'})
	}

	// write cmd
	buklCmd(cmd, buf)

	for i := 0; i < len(args); i++ {
		buklCmd(args[i], buf)
	}
	return buf.Bytes(), len(buf.Bytes())
}

type masterInfo struct {
	id string
	ip string
}

func (i *masterInfo) String() string {
	return strings.Join([]string{i.id, i.ip}, ",")
}

// Find the current topology based on the command
// And return the cropping information of the current topology
func ProbeTopology(pwd string, mode Mode, addrs ...string) (interface{}, error) {
	var redisClient *c.Client
	defer redisClient.Close()

	if reached, err := Ping(addrs...); err != nil {
		return nil, err
	} else if len(reached) < 1 {
		return nil, errors.New("all addresses are unreachable")
	} else {
		if len(pwd) > 0 {
			redisClient = c.NewClient(reached[0], c.DialPassword(pwd))
		} else {
			redisClient = c.NewClient(reached[0])
		}
	}
	var info interface{}

	switch mode {
	case SentinelMode:
		ss, err := c.Strings(redisClient.Do("sentinel", "master", "mymaster"))
		if err != nil {
			return nil, err
		}
		info = ss
	case ClusterMode:
		s, err := c.String(redisClient.Do("cluster", "nodes"))
		if err != nil {
			return nil, err
		}
		info = s
	case SingleMode:
		info = addrs[0]
	}
	return info, nil
}

func parseCmdReplyToClusterNode(info interface{}) (map[*keyList][]NodeInfo, error) {
	var res = make(map[*keyList][]NodeInfo)

	line, ok := info.(string)
	if !ok {
		return nil, errors.New("the info that needs to be Parse is not the type string is needed")
	}

	ss := strings.Split(line, "\n")
	length := len(ss) - 1
	if length < 1 {
		return nil, errors.New("Parsed cmd result into empty")
	}

	for i := 0; i < length; i++ {
		infoStr := ss[i]
		if !strings.Contains(infoStr, MasterStr) {
			continue
		}

		infoSS := strings.Split(infoStr, " ")
		if len(infoSS) < 4 {
			return nil, errors.New("Parse cluster mode line info error")
		}

		ni := NodeInfo{
			Id:       infoSS[0],
			Addr:     strings.Split(string(infoSS[1]), "@")[0],
			IsMaster: true,
		}
		kl := createkeyList().add(ni.Id)
		if _, exist := res[kl]; !exist {
			res[kl] = make([]NodeInfo, 0)
		}
		res[kl] = append(res[kl], ni)
	}

	for k := range res {
		for i := 0; i < length; i++ {
			infoStr := ss[i]
			if !strings.Contains(infoStr, SlaveStr) {
				continue
			}
			infoSS := strings.Split(infoStr, " ")
			if len(infoSS) < 4 {
				return nil, errors.New("Parse cluster mode line info error")
			}
			if !k.include(infoSS[3]) {
				continue
			}
			id := infoSS[0]
			k.add(id)
			res[k] = append(res[k], NodeInfo{
				Id:       id,
				Addr:     strings.Split(string(infoSS[1]), "@")[0],
				IsMaster: false,
			})
		}
	}

	return res, nil
}

func ParseInfoForMasters(m Mode, info interface{}) ([]string, error) {
	var res []string
	switch m {
	case ClusterMode:
		line, ok := info.(string)
		if !ok {
			return nil, errors.New("The info that needs to be Parse is not the type string is needed")
		}
		/*
			# this 3 master and  3 slave cluster model info
			# need beautify cropping to [$RunID,$realIPaddress]
			cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
			c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
			885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
			656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
			a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
			62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected
		*/

		ss := strings.Split(line, "\n")
		length := len(ss) - 1
		if length < 1 {
			return nil, errors.New("Parsed cmd result into empty")
		}
		res = make([]string, 0)
		for i := 0; i < length; i++ {
			infoStr := ss[i]
			if !strings.Contains(infoStr, MasterStr) {
				// just collect master info
				continue
			}
			if strings.Contains(infoStr, "fail") {
				// just collect non fail master info
				continue
			}
			infoSS := strings.Split(infoStr, " ")
			if len(infoSS) < 4 {
				return nil, errors.New("Parse cluster mode line info error")
			}

			tmpIp := infoSS[1]
			infoIp := strings.Split(string(tmpIp), "@")[0]
			info := &masterInfo{
				id: string(infoSS[0]),
				ip: infoIp,
			}
			res = append(res, info.String())
		}

	case SentinelMode:
		masterSS, ok := info.([]string)
		if !ok {
			return nil, errors.New("The info that needs to be Parse is not the type []string is needed")
		}
		/*
			# this 1 master sentinel model
			# need beautify cropping to [$Runid,$realIPaddress]

			 1) "name"
			 2) "mymaster"
			 3) "ip"
			 4) "10.1.1.228"
			 5) "port"
			 6) "8003"
			 7) "runid"
			 8) "8be9401d095b9072b2e67c59fea68894e14da193"
			 9) "flags"
			10) "master"
			11) "link-pending-commands"
			....
			....
		*/
		m := sliceStr2Dict(masterSS)
		infoStr := fieldSplicing(m, "runid", "ip", "port")
		ss := strings.Split(infoStr, ",")
		info := &masterInfo{
			id: ss[0],
			ip: ss[1] + ":" + ss[2],
		}
		res = []string{info.String()}

	case SingleMode:
		addr, ok := info.(string)
		if !ok {
			return nil, errors.New("The info that needs to be Parse is not the type []string is needed")
		}
		res = []string{addr}
	}

	return res, nil
}

func sliceStr2Dict(ss []string) map[string]string {
	res := make(map[string]string)
	for i := 0; i < len(ss); i = i + 2 {
		res[ss[i]] = ss[i+1]
	}
	return res
}

func fieldSplicing(m map[string]string, cols ...string) string {
	ss := make([]string, len(cols), len(cols))
	for i := 0; i < len(cols); i++ {
		val := ""
		if _, exits := m[cols[i]]; !exits {
			continue
		} else {
			val = m[cols[i]]
		}
		ss[i] = val
	}
	return strings.Join(ss, ",")
}

func ProbeNode(addr string, pwd string) (map[sectionType]map[string]string, error) {
	var redisClient *c.Client
	if len(pwd) < 1 {
		redisClient = c.NewClient(addr)
	} else {
		redisClient = c.NewClient(addr, c.DialPassword(pwd))
	}
	line, err := c.String(redisClient.Do("info"))
	if err != nil {
		return nil, err
	}
	return ParseNodeInfo(line), nil
}

// ParseNodeInfo parse the bukl string returned by the redis info command
func ParseNodeInfo(line string) map[sectionType]map[string]string {
	redisInfo := make(map[sectionType]map[string]string)
	strList := strings.Split(line, "\n")
	selection := ""
	for i, _ := range strList {
		line := strings.TrimSpace(strList[i])
		if line == "" || len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "#") {
			selection = strings.TrimSpace(line[1:])
			redisInfo[sectionType(selection)] = make(map[string]string)
			continue
		}
		contentList := strings.Split(line, ":")
		redisInfo[sectionType(selection)][contentList[0]] = contentList[1]
	}
	return redisInfo
}

// parsing information about the replication selection
func ParseReplicationInfo(m map[string]string) map[string]string {
	/*
		role:master
		connected_slaves:1
		slave0:ip=10.1.1.228,port=7004,state=online,offset=3689968249,lag=1
		master_replid:17270cf205f7c98c4c8e80c348fd0564132e6643
		master_replid2:0000000000000000000000000000000000000000
		...
		...
	*/
	if len(m) < 1 {
		return nil
	}
	slaveReg, _ := regexp.Compile("^slave([0-9]*)")
	slaveMapping := make(map[string]string)

	tmpInfolines := make([]string, 0)
	for key, value := range m {
		if !slaveReg.MatchString(key) {
			continue
		}
		infoss := strings.Split(value, ",")
		if len(infoss) < 2 {
			return nil
		}
		for _, info := range infoss {
			// ip=10.1.1.228,...
			infoLine := strings.Split(info, "=")
			tmpInfolines = append(tmpInfolines, infoLine...)
		}
		args := sliceStr2Dict(tmpInfolines)
		slaveMapping[key] =
			strings.Join(
				strings.Split(fieldSplicing(args, "ip", "port"), ","),
				":",
			)
	}
	return slaveMapping
}

// ParseSlaveInfo call the probeNodeInfo function to implement the NodeInfo initialization parameters
func ParseSlaveInfo(s map[string]string, pwd string) ([]*NodeInfo, error) {
	/*
		map[string]string{"slave0":"10.1.1.228:7004"}
	*/
	res := make([]*NodeInfo, 0)
	for _, v := range s {
		allInfoMap, err := ProbeNode(v, pwd)
		if err != nil {
			return nil, err
		}
		selectionServerMap, exist := allInfoMap[Server]
		if !exist {
			return nil, errors.New("probe info error")
		}
		version, exist := selectionServerMap["redis_version"]
		if !exist {
			return nil, errors.New("selection Server.redis_version not exist")
		}

		runid, exist := selectionServerMap["run_id"]
		if !exist {
			return nil, errors.New("selection Server.run_id not exist")
		}

		res = append(res, &NodeInfo{
			Ver: version, Id: runid, Addr: v,
		})
	}
	return res, nil
}
