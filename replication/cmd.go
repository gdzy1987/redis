package replication

import (
	"bytes"
	"strings"

	"github.com/dengzitong/redis/hack"
)

// Command all the command combinations
type Command struct {
	ctype CommandType
	data  [][]byte
}

func (c Command) String() string {
	dlength := len(c.data)
	ss := make([]string, dlength, dlength)
	for i := range c.data {
		ss[i] = hack.String(c.data[i])
	}
	return strings.Join(ss, " ")
}

func (c Command) Data() []byte {
	buf := bytes.NewBuffer(nil)
	for i := range c.data {
		buf.Write(c.data[i])
	}
	return buf.Bytes()
}

func (c Command) Type() CommandType {
	cmdStr := hack.String(c.data[0])
	t, exist := CommandTypeMap[cmdStr]
	if !exist {
		return Undefined
	}
	return t
}

func (c Command) Name() string {
	return CommandNameMap[c.ctype]
}

func (c Command) Arguments() []interface{} {
	length := len(c.data) - 1
	args := make([]interface{}, length, length)
	for i := range c.data[1:] {
		args[i] = c.data[i+1]
	}
	return args
}

func NewCommand(cmdType CommandType, args ...[]byte) Command {
	return Command{
		ctype: cmdType,
		data:  args,
	}
}
