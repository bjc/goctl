package goctl

import (
	"fmt"
	"sort"
	"strings"
)

type cmdHelp struct{}

func (cmd cmdHelp) Name() string {
	return "help"
}

func (cmd cmdHelp) Help() string {
	return "show this message"
}

func (cmd cmdHelp) Run(gc *Goctl, args []string) string {
	rc := []string{"Available commands:", ""}
	cmds := []string{}
	for _, h := range gc.handlers {
		cmds = append(cmds, fmt.Sprintf("\t%s\t%s", h.Name(), h.Help()))
	}
	sort.Strings(cmds)
	rc = append(rc, cmds...)
	return strings.Join(rc, "\n")
}

func init() {
	builtinHandlers = append(builtinHandlers, cmdHelp{})
}
