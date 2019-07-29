package main

import (
	"fmt"
	"os"

	r "github.com/dengzitong/redis/replication"
	t "github.com/dengzitong/redis/topology"
)

type outerImpl struct{}

func (o *outerImpl) Receive(c r.Command) {
	fmt.Printf("command type = %s , args = %s\n", c.Type(), c.Paramters())
}

func main() {
	tp, err := t.NewTopologyist(t.SingleMode, "", "127.0.0.1:6379")
	if err != nil {
		panic(err)
	}

	repl := r.NewReplication(tp, os.Stdout, &outerImpl{})

	cancel, err := repl.Start()
	defer cancel()
	if err != nil {
		panic(err)
	}

}
