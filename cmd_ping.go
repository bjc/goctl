package goctl

type cmdPing struct{}

func (cmd *cmdPing) Name() string {
	return "ping"
}

func (cmd *cmdPing) Run(_ []string) string {
	return "pong"
}

func init() {
	builtinHandlers = append(builtinHandlers, &cmdPing{})
}
