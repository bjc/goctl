package goctl

import (
	"net"
	"reflect"
	"testing"
)

const SOCKPATH = "/tmp/xmppbot"

func TestPing(t *testing.T) {
	gc := NewGoctl(SOCKPATH)
	if err := gc.Start(); err != nil {
		t.Fatalf("Couldn't start: %s.", err)
	}
	defer gc.Stop()

	c, err := net.Dial("unix", SOCKPATH)
	if err != nil {
		t.Fatalf("Couldn't open %s: %s.", SOCKPATH, err)
	}
	defer c.Close()

	buf := []byte("ping")
	Write(c, buf)

	buf, err = Read(c)
	if err != nil {
		t.Fatalf("Couldn't read from socket: %s.", err)
	}
	s := string(buf)
	if s != "pong" {
		t.Errorf("Sending ping: got '%s', expected 'pong'.", s)
	}
}

func TestAlreadyRunning(t *testing.T) {
	gc := NewGoctl(SOCKPATH)
	if err := gc.Start(); err != nil {
		t.Errorf("Couldn't start: %s.", err)
	}
	defer gc.Stop()

	bar := NewGoctl(SOCKPATH)
	if bar.Start() == nil {
		t.Errorf("Started bot when already running.")
	}
	defer bar.Stop()
}

func TestSimulCommands(t *testing.T) {
	gc := NewGoctl(SOCKPATH)
	if err := gc.Start(); err != nil {
		t.Fatalf("Couldn't start: %s.", err)
	}
	defer gc.Stop()

	c0, err := net.Dial("unix", SOCKPATH)
	if err != nil {
		t.Fatalf("Coudln't open first connection to %s: %s.", SOCKPATH, err)
	}
	defer c0.Close()

	c1, err := net.Dial("unix", SOCKPATH)
	if err != nil {
		t.Fatalf("Coudln't open second connection to %s: %s.", SOCKPATH, err)
	}
	defer c1.Close()

	Write(c0, []byte("ping"))
	Write(c1, []byte("ping"))

	if buf, err := Read(c0); err != nil {
		t.Errorf("Couldn't read from first connection: %s.", err)
	} else if string(buf) != "pong" {
		t.Fatalf("Sending ping on first connection: got '%s', expected 'pong'.", string(buf))
	}
	if buf, err := Read(c1); err != nil {
		t.Fatalf("Couldn't read from second connection: %s.", err)
	} else if string(buf) != "pong" {
		t.Fatalf("Sending ping on second connection: got '%s', expected 'pong'.", string(buf))
	}
}

func TestAddHandler(t *testing.T) {
	gc := NewGoctl(SOCKPATH)
	if err := gc.Start(); err != nil {
		t.Fatalf("Couldn't start: %s", err)
	}
	defer gc.Stop()

	gc.AddHandler("foo", func(args []string) string {
		if !reflect.DeepEqual(args, []string{"bar", "baz"}) {
			t.Errorf("Got %v, expected ['bar', 'baz']", args)
		}
		return "bar baz"
	})

	c, err := net.Dial("unix", SOCKPATH)
	if err != nil {
		t.Fatalf("Coudln't open connection to %s: %s.", SOCKPATH, err)
	}
	defer c.Close()
	Write(c, []byte("foo bar baz"))

	if buf, err := Read(c); err != nil {
		t.Errorf("Couldn't read from connection: %s.", err)
	} else if string(buf) != "bar baz" {
		t.Errorf("Got: %s, expected 'bar baz'")
	}
}

func TestAddHandlers(t *testing.T) {
	gc := NewGoctl(SOCKPATH)
	if err := gc.Start(); err != nil {
		t.Fatalf("Couldn't start: %s", err)
	}
	defer gc.Stop()

	gc.AddHandlers([]*Handler{
		{"foo", func(args []string) string {
			if !reflect.DeepEqual(args, []string{"bar", "baz"}) {
				t.Errorf("Got %v, expected ['bar', 'baz']", args)
			}
			return ""
		}},
		{"bar", func(args []string) string {
			if !reflect.DeepEqual(args, []string{"baz", "pham"}) {
				t.Errorf("Got %v, expected ['baz', 'pham']", args)
			}
			return "wauug"
		}}})

	c, err := net.Dial("unix", SOCKPATH)
	if err != nil {
		t.Fatalf("Coudln't open connection to %s: %s.", SOCKPATH, err)
	}
	defer c.Close()
	Write(c, []byte("foo bar baz"))
	Write(c, []byte("bar baz pham"))

	if buf, err := Read(c); err != nil {
		t.Errorf("Couldn't read from connection: %s.", err)
	} else if string(buf) != "" {
		t.Errorf("Got: %s, expected ''", string(buf))
	}

	if buf, err := Read(c); err != nil {
		t.Errorf("Couldn't read from connection: %s.", err)
	} else if string(buf) != "wauug" {
		t.Errorf("Got: %s, expected 'wauug'", string(buf))
	}
}

func BenchmarkStartStop(b *testing.B) {
	gc := NewGoctl(SOCKPATH)
	for i := 0; i < b.N; i++ {
		gc.Start()
		gc.Stop()
	}
}

func BenchmarkPing(b *testing.B) {
	gc := NewGoctl(SOCKPATH)
	gc.Start()
	defer gc.Stop()

	c, err := net.Dial("unix", SOCKPATH)
	if err != nil {
		b.Fatalf("Coudln't open connection to %s: %s.", SOCKPATH, err)
	}
	defer c.Close()

	for i := 0; i < b.N; i++ {
		Write(c, []byte("ping"))
		Read(c)
	}
}
