package topology

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

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

func Command(strs ...string) ([]byte, int) {
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte{'*'})
	lengthStr := fmt.Sprintf("%d", len(strs))
	buf.Write([]byte(lengthStr))
	buf.Write([]byte{'\r', '\n'})
	for _, str := range strs {
		buf.Write([]byte("$"))
		buf.Write([]byte(fmt.Sprintf("%d", len(str))))
		buf.Write([]byte{'\r', '\n'})
		buf.Write([]byte(str))
		buf.Write([]byte{'\r', '\n'})
	}
	return buf.Bytes(), len(buf.Bytes())
}
