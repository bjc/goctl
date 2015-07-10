package goctl

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

// How long to wait for a response before returning an error.
const timeout = 100 * time.Millisecond

var builtinHandlers []Handler

// Error returned if handler already exists.
var HandlerExists error

var (
	lastid uint64
	Logger log15.Logger
)

type Goctl struct {
	logger   log15.Logger
	path     string
	listener net.Listener
	handlers map[string]Handler
}

type Handler interface {
	Name() string
	Run([]string) string
}

func init() {
	Logger = log15.New()
	Logger.SetHandler(log15.DiscardHandler())

	HandlerExists = errors.New("handler exists")
}

func NewGoctl(path string) Goctl {
	handlers := make(map[string]Handler)
	for _, h := range builtinHandlers {
		handlers[h.Name()] = h
	}
	return Goctl{
		logger:   Logger.New("id", atomic.AddUint64(&lastid, 1)),
		path:     path,
		handlers: handlers,
	}
}

func (gc *Goctl) Start() error {
	gc.logger.Info("Starting command listener.")

	if pid := gc.isAlreadyRunning(); pid != "" {
		gc.logger.Crit("Command listener already running.", "pid", pid)
		return fmt.Errorf("already running on pid %s", pid)
	} else {
		os.Remove(gc.path)
	}

	var err error
	gc.listener, err = net.Listen("unix", gc.path)
	if err != nil {
		gc.logger.Crit("Couldn't listen on socket.", "path", gc.path, "error", err)
		return err
	}

	go gc.acceptor()
	return nil
}

func (gc *Goctl) Stop() {
	gc.logger.Info("Stopping command listener.")

	if gc.listener != nil {
		gc.listener.Close()
		os.Remove(gc.path)
	}
}

func (gc *Goctl) AddHandler(h Handler) error {
	return gc.AddHandlers([]Handler{h})
}

func (gc *Goctl) AddHandlers(handlers []Handler) error {
	for _, h := range handlers {
		if gc.handlers[h.Name()] != nil {
			return HandlerExists
		}
		gc.handlers[h.Name()] = h
	}
	return nil
}

func Read(r io.Reader) ([]byte, error) {
	var n uint16
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, err
	}

	p := make([]byte, n)
	if n == 0 {
		// Don't attempt to read no data or EOF will be
		// returned. Instead, just return an empty byte slice.
		return p, nil
	}

	if _, err := r.Read(p[:]); err != nil {
		return nil, err
	}
	return p, nil
}

func Write(w io.Writer, p []byte) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(p))); err != nil {
		return err
	}
	if _, err := w.Write(p); err != nil {
		return err
	}
	return nil
}

func (gc *Goctl) acceptor() {
	for {
		c, err := gc.listener.Accept()
		if err != nil {
			gc.logger.Error("Error accepting connection.", "error", err)
			return
		}
		go gc.reader(c)
	}
}

func (gc *Goctl) reader(c io.ReadWriteCloser) error {
	defer gc.logger.Info("Connection closed.")
	defer c.Close()

	gc.logger.Info("New connection.")
	for {
		buf, err := Read(c)
		if err != nil {
			gc.logger.Error("Error reading from connection.", "error", err)
			return err
		}

		cmd := strings.Split(string(buf), "\u0000")
		gc.logger.Debug("Got command.", "cmd", cmd)
		var resp string
		if h := gc.handlers[cmd[0]]; h != nil {
			resp = h.Run(cmd[1:])
		} else {
			resp = fmt.Sprintf("ERROR: unknown command: '%s'.", cmd[0])
		}
		gc.logger.Debug("Responding.", "resp", resp)
		Write(c, []byte(resp))
	}
	/* NOTREACHED */
}

func (gc *Goctl) isAlreadyRunning() string {
	c, err := net.Dial("unix", gc.path)
	if err != nil {
		return ""
	}
	defer c.Close()

	dataChan := make(chan []byte, 1)
	c.Write([]byte("pid"))
	go func() {
		buf := make([]byte, 1024)
		n, err := c.Read(buf[:])
		if err == nil {
			dataChan <- buf[0:n]
		}
	}()

	select {
	case buf := <-dataChan:
		return string(buf)
	case <-time.After(timeout):
		gc.logger.Info("Timed out checking PID of existing service.", "path", gc.path)
		return ""
	}
}
