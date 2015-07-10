package goctl

import (
	"os"
	"strconv"
)

var pid string

type cmdPID struct{}

func (cmd cmdPID) Name() string {
	return "pid"
}

func (cmd cmdPID) Help() string {
	return "return the Unix process ID of this program"
}

func (cmd cmdPID) Run(_ *Goctl, _ []string) string {
	return pid
}

func init() {
	pid = strconv.Itoa(os.Getpid())
	builtinHandlers = append(builtinHandlers, cmdPID{})
}
