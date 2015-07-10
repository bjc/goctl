package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/bjc/goctl"
)

func main() {
	var path = flag.String("f", "", "socket path for sending commands (required)")
	flag.Parse()

	if *path == "" {
		flag.Usage()
		os.Exit(1)
	}

	c, err := net.Dial("unix", *path)
	if err != nil {
		log.Fatalf("Couldn't connect to %s: %s.", *path, err)
	}
	defer c.Close()

	goctl.Write(c, []byte(strings.Join(flag.Args(), "\u0000")))
	if buf, err := goctl.Read(c); err != nil {
		log.Fatalf("Error reading response from command: %s.", err)
	} else {
		fmt.Println(string(buf))
	}
}
