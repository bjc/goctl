package goctl

import (
	"strconv"
	"testing"
)

func TestPID(t *testing.T) {
	gc := start(t)
	defer gc.Stop()

	c := dial(t)
	defer c.Close()

	buf := []byte("pid")
	Write(c, buf)

	buf, err := Read(c)
	if err != nil {
		t.Fatalf("Couldn't read from socket: %s.", err)
	}
	if _, err = strconv.Atoi(string(buf)); err != nil {
		t.Errorf("Requesting PID: got non-integer: '%s'.", string(buf))
	}
}
