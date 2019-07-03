package topology

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	c "github.com/dengzitong/redis/client"
)

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

	bulkCmd := func(v interface{}, buf *bytes.Buffer) {
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
	bulkCmd(cmd, buf)

	for i := 0; i < len(args); i++ {
		bulkCmd(args[i], buf)
	}
	return buf.Bytes(), len(buf.Bytes())
}

type info struct {
	id       string
	ip       string
	isMaster bool
	masterId string
}

func (i *info) String() string {
	var master string = "0"
	if i.isMaster {
		master = "1"
	}
	if len(i.masterId) == 0 {
		i.masterId = "nil"
	}
	return strings.Join([]string{i.id, i.ip, master, i.masterId}, ",")
}

// Find the current topology based on the command
// And return the cropping information of the current topology
func ProbeTopology(mode Mode, rcli RClient, addrs ...string) ([]string, error) {
	var res []string

	switch mode {
	case ClusterMode:
		line, err := c.String(rcli.Do("cluster", "nodes"))
		if err != nil {
			return nil, err
		}
		/*
			# this 3 master and  3 slave cluster model info
			# need beautify cropping to [$RunID,$realIPaddress,$Role,$MasterRunID]
			cebd9205cbde0d1ec4ad75600849a88f1f6294f6 10.1.1.228:7005@17005 master - 0 1562154209390 32 connected 5461-10922
			c6d165b72cfcd76d7662e559dc709e00e3dabf03 10.1.1.228:7001@17001 myself,master - 0 1562154207000 25 connected 0-5460
			885493415bea22919fc9ce83836a9e6a8d0c1314 10.1.1.228:7003@17003 master - 0 1562154207000 24 connected 10923-16383
			656042ad560b887164138a19dab2502154f8b039 10.1.1.228:7004@17004 slave c6d165b72cfcd76d7662e559dc709e00e3dabf03 0 1562154205381 25 connected
			a70fbd191b4e00ff6d65c71d9d2c6f15d1adbcab 10.1.1.228:7002@17002 slave cebd9205cbde0d1ec4ad75600849a88f1f6294f6 0 1562154208000 32 connected
			62bd020a2a5121a27c0e5540d1f0d4bba08cebb2 10.1.1.228:7006@17006 slave 885493415bea22919fc9ce83836a9e6a8d0c1314 0 1562154208388 24 connected
		*/
		ss := strings.Split(line, "\n")
		length := len(ss)
		if length < 1 {
			return nil, errors.New("probe execute cmd result empty")
		}
		res = make([]string, length, length)
		for i := 0; i < length; i++ {
			infoStr := ss[i]
			infoSS := strings.Split(infoStr, " ")
			if len(infoSS) < 4 {
				return nil, errors.New("parsed cluster mode line info error")
			}

			var isMaster bool
			if strings.Contains(infoStr, MasterStr) {
				isMaster = true
			}

			var masterId string
			if isMaster {
				masterId = string(infoStr[3])
			}

			info := &info{
				id:       string(infoStr[0]),
				ip:       string(infoStr[1]),
				isMaster: isMaster,
				masterId: masterId,
			}
			res[i] = info.String()
		}
	case SentinelMode:
		masterSS, err := c.Strings(rcli.Do("sentinel", "master", "mymaster"))
		if err != nil {
			return nil, err
		}
		m := sliceStr2Dict(masterSS)
		infoStr := fieldSplicing(m, "runid", "ip", "port")
		ss := strings.Split(infoStr, ",")
		info := &info{
			id:       ss[0],
			ip:       ss[1] + ":" + ss[2],
			isMaster: true,
			masterId: "",
		}
		res = []string{info.String()}
		// 要不要继续拿slave的地址
	case SingleMode:

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
