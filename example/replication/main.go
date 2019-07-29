package main

import (
	"os"

	r "github.com/dengzitong/redis/replication"
	t "github.com/dengzitong/redis/topology"
)

func main() {
	tp, err := t.NewTopologyist(t.SingleMode, "", "127.0.0.1:6379")
	if err != nil {
		panic(err)
	}

	repl := r.NewReplication(tp, os.Stdout)

	cancel, err := repl.Start()
	defer cancel()
	if err != nil {
		panic(err)
	}

}
