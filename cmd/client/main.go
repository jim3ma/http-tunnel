package main

import (
	"fmt"
	"io"
	"time"

	"github.com/jim3ma/http-tunnel"
)

func main() {
	h := ht.NewTunnel("http", "127.0.0.1", "10080", "/ping", "/pong")
	conn, err := h.Dial("tcp", "127.0.0.1:10080")
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	go func() {
		for i := 0; i < 10; i++ {
			conn.Write([]byte("test"))
			time.Sleep(time.Second)
		}
		//conn.Close()
	}()
	size := 32 * 1024
	buf := make([]byte, size)
	for {
		n, err := conn.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Printf("%v", err)
			break
		}
		if err == io.EOF {
			fmt.Printf("%v", err)
			break
		}
		fmt.Printf("%s\n", string(buf[0:n]))
	}
}
