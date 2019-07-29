package replication

import (
	"bytes"
)

// Command all the command combinations
type Command struct {
	ctype CommandType
	data  [][]byte
}

func (c Command) Data() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte(CommandNameMap[c.ctype]))
	for i := range c.data {
		buf.Write(c.data[i])
	}
	return buf.Bytes()
}

func (c Command) Type() []byte {
	return []byte(CommandNameMap[c.ctype])
}

func (c Command) Name() string {
	return CommandNameMap[c.ctype]
}

func (c Command) Paramters() [][]byte {
	return c.data
}

func NewCommand(cmdType CommandType, args ...[]byte) Command {
	return Command{
		ctype: cmdType,
		data:  args,
	}
}
