package goctl

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"reflect"
	"testing"
)

type testHandler struct {
	name string
	fn   func(*Goctl, []string) string
}

func makeHandler(name string, fn func(_ *Goctl, _ []string) string) testHandler {
	return testHandler{name: name, fn: fn}
}

func (th testHandler) Name() string {
	return th.name
}

func (th testHandler) Help() string {
	return ""
}

func (th testHandler) Run(gc *Goctl, args []string) string {
	return th.fn(gc, args)
}

var sockpath string

func init() {
	f, err := ioutil.TempFile("", "goctl-test")
	if err != nil {
		log.Fatalf("Couldn't create temporary file: %s.", err)
	}
	sockpath = f.Name()
	if err := os.Remove(sockpath); err != nil {
		log.Fatalf("Couldn't delete temporary file '%s': %s.", sockpath, err)
	}
}

func start(t testing.TB) *Goctl {
	gc := NewGoctl(sockpath)
	if err := gc.Start(); err != nil {
		t.Fatalf("Couldn't start: %s.", err)
	}
	return &gc
}

func dial(t testing.TB) net.Conn {
	c, err := net.Dial("unix", sockpath)
	if err != nil {
		t.Fatalf("Couldn't open %s: %s.", sockpath, err)
	}
	return c
}

func TestAlreadyRunning(t *testing.T) {
	gc := start(t)
	defer gc.Stop()

	bar := NewGoctl(sockpath)
	if bar.Start() == nil {
		t.Errorf("Started bot when already running.")
	}
	defer bar.Stop()
}

func TestSimulCommands(t *testing.T) {
	gc := start(t)
	defer gc.Stop()

	c0 := dial(t)
	defer c0.Close()

	c1 := dial(t)
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
	gc := start(t)
	defer gc.Stop()

	gc.AddHandler(makeHandler("foo", func(innerGC *Goctl, args []string) string {
		if innerGC != gc {
			t.Errorf("Goctl object not passed into handler properly (got: %p, want: %p).", innerGC, gc)
		}
		if !reflect.DeepEqual(args, []string{"bar", "baz"}) {
			t.Errorf("Got %v, expected ['bar', 'baz']", args)
		}
		return "bar baz"
	}))

	c := dial(t)
	defer c.Close()

	Write(c, []byte("foo\u0000bar\u0000baz"))

	if buf, err := Read(c); err != nil {
		t.Errorf("Couldn't read from connection: %s.", err)
	} else if string(buf) != "bar baz" {
		t.Errorf("Got: %s, expected 'bar baz'")
	}
}

func TestAddHandlers(t *testing.T) {
	gc := start(t)
	defer gc.Stop()

	gc.AddHandlers([]Handler{
		makeHandler("foo", func(_ *Goctl, args []string) string {
			if !reflect.DeepEqual(args, []string{"bar", "baz"}) {
				t.Errorf("Got %v, expected ['bar', 'baz']", args)
			}
			return ""
		}),
		makeHandler("bar", func(_ *Goctl, args []string) string {
			if !reflect.DeepEqual(args, []string{"baz", "pham"}) {
				t.Errorf("Got %v, expected ['baz', 'pham']", args)
			}
			return "wauug"
		}),
	})

	c := dial(t)
	defer c.Close()

	Write(c, []byte("foo\u0000bar\u0000baz"))
	Write(c, []byte("bar\u0000baz\u0000pham"))

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

func TestCannotOverrideExtantHandlers(t *testing.T) {
	gc := start(t)
	defer gc.Stop()

	err := gc.AddHandler(makeHandler("ping", func(_ *Goctl, args []string) string {
		return "gnip"
	}))
	if err != HandlerExists {
		t.Error("Was able to override built-in ping handler.")
	}
	err = gc.AddHandlers([]Handler{
		makeHandler("foo", func(_ *Goctl, args []string) string { return "foo" }),
		makeHandler("foo", func(_ *Goctl, args []string) string { return "foo" }),
	})
	if err != HandlerExists {
		t.Error("Was able to assign two handlers for 'foo'.")
	}
}

func BenchmarkStartStop(b *testing.B) {
	gc := NewGoctl(sockpath)
	for i := 0; i < b.N; i++ {
		gc.Start()
		gc.Stop()
	}
}

func BenchmarkPing(b *testing.B) {
	gc := start(b)
	defer gc.Stop()

	c := dial(b)
	defer c.Close()

	for i := 0; i < b.N; i++ {
		Write(c, []byte("ping"))
		Read(c)
	}
}
