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
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

var (
	flagServer   = flag.String("server", "", "Server to connect to")
	flagKey      = flag.String("key", "", "Key to use")
	flagSocks5   = flag.String("socks5", "", "Connect through SOCKS5 proxy")
	flagDeadline = flag.Duration("deadline", time.Minute*5, "Deadline of connections")
)

func main() {
	flag.Parse()

	failedToConnect := 0

	for {
		var (
			conn net.Conn
			err  error
		)
		if *flagSocks5 == "" {
			conn, err = net.Dial("tcp", *flagServer)
			if err != nil {
				log.Printf("error connecting to server: %v", err)
				failedToConnect += 1
				goto fail
			} else {
				failedToConnect = 0
			}
		} else {
			dialer, err := proxy.SOCKS5("tcp", *flagSocks5, nil, &net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 15 * time.Second,
			})
			if err != nil {
				log.Printf("error connecting to proxy: %v", err)
				goto fail
			}
			conn, err = dialer.Dial("tcp", *flagServer)
			if err != nil {
				log.Printf("error connecting to server: %v", err)
				failedToConnect += 1
				goto fail
			} else {
				failedToConnect = 0
			}
		}

		conn.SetDeadline(time.Now().Add(*flagDeadline))
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

				deadline := time.Now().Add(*flagDeadline)
				conn.SetDeadline(deadline)
				local.SetDeadline(deadline)

				wg := new(sync.WaitGroup)
				wg.Add(2)
				go func(wg *sync.WaitGroup, local, conn net.Conn) {
					defer wg.Done()
					io.Copy(local, conn)
				}(wg, local, conn)
				go func(wg *sync.WaitGroup, local, conn net.Conn) {
					defer wg.Done()
					io.Copy(conn, local)
				}(wg, local, conn)
				go func(wg *sync.WaitGroup) {
					wg.Wait()
					local.Close()
					conn.Close()
					log.Println("client connection closed")
				}(wg)

				goto success
			}
		}

	fail:
		if conn != nil {
			conn.Close()
		}
		if failedToConnect >= 5 {
			// let supervisor restart this process
			log.Fatalf("too many reconnections, trying restart program")
		}
		time.Sleep(5 * time.Second)

	success:
	}
}
