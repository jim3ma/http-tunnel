package ht

import (
	"net/http"
	"sync"
	"net"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

const connectionUidLength = 36

var connections sync.Map

func init() {
	connections = sync.Map{}
}

func pushConnection(uid string, obj interface{}) {
	connections.Store(uid, obj)
}

func popConnection(uid string) (net.Conn, error) {
	conn, ok := connections.Load(uid)
	if !ok {
		return nil, fmt.Errorf("connection not found")
	}
	connections.Delete(uid)
	return conn.(net.Conn), nil
}

func Phase1Handler(w http.ResponseWriter, r *http.Request) {
	request, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	data := strings.Split(string(request), "\n")
	if err != nil || len(data) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uid := data[0]
	addr := data[1]
	conn, err := net.Dial("tcp", string(addr))
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	pushConnection(uid, conn)

	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.(http.Flusher).Flush()

	ioCopy(w, conn, func() {
		w.(http.Flusher).Flush()
	})
}

func Phase2Handler(w http.ResponseWriter, r *http.Request) {
	uid := make([]byte, connectionUidLength)
	n, err := r.Body.Read(uid)
	defer r.Body.Close()
	if err != nil || n != connectionUidLength {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := popConnection(string(uid))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = ioCopy(conn, r.Body, nil)
	if err != nil && err != io.EOF {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func ioCopy(dst io.Writer, src io.Reader, action func()) (written int64, err error) {
	size := 32 * 1024
	buf := make([]byte, size)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if action != nil {
				action()
			}
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}
