package topology

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
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
