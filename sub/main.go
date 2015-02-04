package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	flagServer = flag.String("server", "", "Server to connect to")
	flagKey    = flag.String("key", "", "Key to use")
)

func main() {
	flag.Parse()

	for {
		conn, err := net.Dial("tcp", *flagServer)
		if err != nil {
			log.Printf("error connecting to server: %v", err)
			goto fail
		}
		conn.SetDeadline(time.Now().Add(2 * time.Minute))
		if _, err := fmt.Fprintf(conn, "SUBSCRIBE %s\n", *flagKey); err != nil {
			goto fail
		}
		log.Println("connected")
		{
			r := bufio.NewReader(conn)
			line, err := r.ReadString('\n')
			if err != nil {
				log.Println("error reading:", err)
				goto fail
			}
			parts := strings.Split(strings.TrimSpace(line), " ")
			switch strings.ToUpper(parts[0]) {
			case "BIND":
				if len(parts) < 2 {
					goto fail
				}
				port, err := strconv.Atoi(parts[1])
				if err != nil {
					log.Println(err)
					goto fail
				}
				local, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", port))
				if err != nil {
					log.Println(err)
					goto fail
				}
				conn.SetDeadline(time.Time{})
				go io.Copy(local, conn)
				go io.Copy(conn, local)
				goto success
			}
		}
	fail:
		if conn != nil {
			conn.Close()
		}
		time.Sleep(5 * time.Second)
	success:
	}
}
