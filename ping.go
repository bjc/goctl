package goctl

type cmdPing struct{}

func (cmd cmdPing) Name() string {
	return "ping"
}

func (cmd cmdPing) Help() string {
	return "checks whether the connection is working"
}

func (cmd cmdPing) Run(_ *Goctl, _ []string) string {
	return "pong"
}

func init() {
	builtinHandlers = append(builtinHandlers, &cmdPing{})
}
