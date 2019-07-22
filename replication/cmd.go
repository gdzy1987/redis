package replication

import (
	"strings"

	"github.com/dengzitong/redis/hack"
)

// Command all the command combinations
type Command struct {
	T CommandType
	D [][]byte
}

func (c *Command) String() string {
	ss := make([]string, len(c.D), len(c.D))
	for i := range c.D {
		ss[i] = hack.String(c.D[i])
	}
	return strings.Join(ss, " ")
}

func (c *Command) Type() CommandType {
	cmdStr := hack.String(c.D[0])
	t, exist := CommandTypeMap[cmdStr]
	if !exist {
		return Undefined
	}
	return t
}

func (c *Command) CmdTypeName() string {
	return CommandNameMap[c.T]
}

func (c *Command) Args() []interface{} {
	args := make([]interface{}, len(c.D)-1, len(c.D)-1)
	for i := range c.D[1:] {
		args[i] = c.D[i+1]
	}
	return args
}

func NewCommand(cmdType CommandType, args ...[]byte) *Command {
	return &Command{
		T: cmdType,
		D: args,
	}
}
