package goctl

import "testing"

func TestPing(t *testing.T) {
	gc := start(t)
	defer gc.Stop()

	c := dial(t)
	defer c.Close()

	buf := []byte("ping")
	Write(c, buf)

	buf, err := Read(c)
	if err != nil {
		t.Fatalf("Couldn't read from socket: %s.", err)
	}
	s := string(buf)
	if s != "pong" {
		t.Errorf("Sending ping: got '%s', expected 'pong'.", s)
	}
}
