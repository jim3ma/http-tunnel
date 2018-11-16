package ht

import (
	"crypto/tls"
	"net"
	"errors"
	"time"
	"net/http"
	"bytes"
	"io"
	"log"
	"net/url"
	"sync"
	"io/ioutil"

	"github.com/satori/go.uuid"
)

type HttpTunnel struct {
	Schema     string
	Address    string
	Port       string
	Phase1Path string
	Phase2Path string
	HttpClient *http.Client
	Userinfo   *url.Userinfo
}

func NewTunnel(schema, address, port string, phase1, phase2 string, user *url.Userinfo) *HttpTunnel {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	return &HttpTunnel{
		Schema:     schema,
		Address:    address,
		Port:       port,
		Phase1Path: phase1,
		Phase2Path: phase2,
		HttpClient: httpClient,
		Userinfo:   user,
	}
}

func (h *HttpTunnel) Dial(network, addr string) (net.Conn, error) {
	switch network {
	case "tcp", "tcp6", "tcp4":
	default:
		return nil, errors.New("proxy: no support for http tunnel proxy connections of type " + network)
	}

	uid, _ := uuid.NewV4()
	log.Printf("uid: %s\n", uid.String())
	id := uid.String()

	r, err := h.phase1(id, addr)
	if err != nil {
		return nil, err
	}
	w := h.phase2(id)

	conn := &httpTunnelConn{
		r: r,
		w: w,
	}
	return conn, nil
}

func (h *HttpTunnel) upstream() string {
	return h.Schema + "://" + h.Address + ":" + h.Port
}

func (h *HttpTunnel) phase1(uid, addr string) (io.ReadCloser, error) {
	u, _ := url.Parse(h.upstream() + h.Phase1Path)
	if h.Userinfo != nil {
		u.User = h.Userinfo
	}
	req := &http.Request{
		Method:           "POST",
		ProtoMajor:       1,
		ProtoMinor:       1,
		URL:              u,
		TransferEncoding: []string{"chunked"},
		Body:             ioutil.NopCloser(bytes.NewBuffer([]byte(uid + "\n" + addr))),
		Header:           make(map[string][]string),
	}

	req.Header.Set("Content-Type", "text/xml")
	resp, err := h.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (h *HttpTunnel) phase2(uid string) io.WriteCloser {
	r, w := io.Pipe()
	u, _ := url.Parse(h.upstream() + h.Phase2Path)
	if h.Userinfo != nil {
		u.User = h.Userinfo
	}
	req := &http.Request{
		Method:           "POST",
		ProtoMajor:       1,
		ProtoMinor:       1,
		URL:              u,
		TransferEncoding: []string{"chunked"},
		Body:             r,
		Header:           make(map[string][]string),
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		w.Write([]byte(uid))
		wg.Done()
	}()

	req.Header.Set("Content-Type", "application/octet-stream")

	go func() {
		resp2, err := h.HttpClient.Do(req)
		if nil != err {
			log.Printf("error => %s", err.Error())
			return
		}
		resp2.Body.Close()
	}()
	wg.Wait()
	return w
}

type httpTunnelConn struct {
	r io.ReadCloser
	w io.WriteCloser
}

func (h *httpTunnelConn) Read(b []byte) (n int, err error) {
	return h.r.Read(b)
}

func (h *httpTunnelConn) Write(b []byte) (n int, err error) {
	return h.w.Write(b)
}

func (h *httpTunnelConn) Close() error {
	h.r.Close()
	h.w.Close()
	return nil
}

func (h *httpTunnelConn) LocalAddr() net.Addr {
	return nil
}

func (h *httpTunnelConn) RemoteAddr() net.Addr {
	return nil
}

func (h *httpTunnelConn) SetDeadline(t time.Time) error {
	return nil
}

func (h *httpTunnelConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (h *httpTunnelConn) SetWriteDeadline(t time.Time) error {
	return nil
}
