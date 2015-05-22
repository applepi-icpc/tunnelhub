package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var (
	flagServer = flag.String("s", "localhost:5555", "Server address")
	flagKey    = flag.String("k", "", "Key")
	flagPort   = flag.Int("p", 22, "Port")
)

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", *flagServer)
	check(err)
	_, err = fmt.Fprintf(conn, "connect %s %d\n", *flagKey, *flagPort)
	check(err)
	go func() {
		io.Copy(conn, os.Stdin)
		conn.Close()
	}()
	io.Copy(os.Stdout, conn)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
