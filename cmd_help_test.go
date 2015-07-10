package goctl

import "testing"

func TestHelp(t *testing.T) {
	gc := start(t)
	defer gc.Stop()

	c := dial(t)
	defer c.Close()

	buf := []byte("help")
	Write(c, buf)

	buf, err := Read(c)
	if err != nil {
		t.Fatalf("Couldn't read from socket: %s.", err)
	}

	got := string(buf)
	want := `Available commands:

	help	show this message
	pid	return the Unix process ID of this program
	ping	checks whether the connection is working`
	if got != want {
		t.Errorf("Didn't get proper help response.\nGot:\n'%s',\nWant:\n'%s'", got, want)
	}
}
