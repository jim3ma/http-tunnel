package main

import (
	"io"
	"time"
	"net/url"
	"net"
	"log"

	"github.com/jim3ma/http-tunnel"
	"fmt"
)

func main() {
	go echo()

	user := url.UserPassword("default", "default-pass")
	h := ht.NewTunnel("http", "127.0.0.1", "10080", "/ping", "/pong", user)
	conn, err := h.Dial("tcp", "127.0.0.1:3333")
	if err != nil {
		log.Printf("%s", err)
		return
	}
	go func() {
		for i := 0; i < 10; i++ {
			conn.Write([]byte(fmt.Sprintf("test %d", i)))
			time.Sleep(time.Second)
		}
		//conn.Close()
	}()
	size := 32 * 1024
	buf := make([]byte, size)
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			log.Printf("%v", err)
			break
		}
		if err == io.EOF {
			log.Printf("%v", err)
			break
		}
		log.Printf("Read new data from server: %s\n", string(buf[0:n]))
	}
}

func echo() {
	l, err := net.Listen("tcp", ":3333")
	if err != nil {
		log.Panicln(err)
	}

	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panicln(err)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	log.Println("Accepted new connection.")
	defer conn.Close()
	defer log.Println("Closed connection.")

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:size]
		//log.Println("Read new data from connection", string(data))
		conn.Write(data)
	}
}