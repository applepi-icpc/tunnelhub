package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	connections = make(map[string]net.Conn)
	actives     = make(map[string]net.Conn)
	mu          sync.Mutex
)

func handle(conn net.Conn) {
	log.Printf("connected: %v", conn.RemoteAddr())

	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Split(strings.TrimSpace(line), " ")
	switch strings.ToUpper(parts[0]) {
	case "SUBSCRIBE":
		if len(parts) < 2 {
			return
		}
		mu.Lock()
		connections[parts[1]] = conn
		mu.Unlock()
		log.Println("SUBCRIBED", parts[1])
	case "LIST":
		mu.Lock()
		fmt.Fprintf(conn, "SUBSCRIBERS:\n")
		for k, c := range connections {
			fmt.Fprintf(conn, "\t%v from %v\n", k, c.RemoteAddr())
		}
		fmt.Fprintf(conn, "ACTIVE CONNECTIONS:\n")
		for k, c := range actives {
			fmt.Fprintf(conn, "\t%v from %v\n", k, c.RemoteAddr())
		}
		mu.Unlock()
		conn.Close()

	case "CONNECT":
		if len(parts) < 3 {
			fmt.Fprintf(conn, "usage: CONNECT token port")
			return
		}
		token := parts[1]
		port, err := strconv.Atoi(parts[2])
		if err != nil {
			fmt.Fprintf(conn, "error parsing port: %v", err)
			return
		}

		mu.Lock()
		client, ok := connections[token]
		if !ok {
			fmt.Fprintf(client, "token not found")
			mu.Unlock()
			return
		}
		delete(connections, token)
		mu.Unlock()

		if _, err := fmt.Fprintf(client, "BIND %d\n", port); err != nil {
			fmt.Fprintf(conn, "error sending BIND request: %v", err)
			client.Close()
		}

		server, err := net.Listen("tcp", "")
		if err != nil {
			fmt.Fprintf(conn, "cannot create listener: %v", err)
			return
		}
		fmt.Fprintf(conn, "listening: %v", server.Addr())
		tokenserver := fmt.Sprintf("%s:%d <==> %s", token, port, server.Addr().String())
		mu.Lock()
		actives[tokenserver] = client
		mu.Unlock()
		go func() {
			defer server.Close()
			conn, err := server.Accept()
			if err != nil {
				return
			}
			go io.Copy(conn, client)
			go func() {
				io.Copy(client, conn)
				mu.Lock()
				delete(actives, token)
				mu.Unlock()
			}()
		}()
		conn.Close()
	default:
		conn.Close()
	}
}

func main() {
	server, err := net.Listen("tcp", ":5555")
	if err != nil {
		log.Panicf("error listening: %v", err)
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Panicf("error accepting: %v", err)
		}
		go handle(conn)
	}
}
