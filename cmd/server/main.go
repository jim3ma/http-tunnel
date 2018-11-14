package main

import (
	"flag"
	"net/http/pprof"
	"net/http"
	"os"
	"fmt"
	"io"

	"github.com/jim3ma/http-tunnel"
)

var (
	phase1Url = flag.String("phase1-path", "/ping", "Phase 1 URL path")
	phase2Url = flag.String("phase2-path", "/pong", "Phase 2 URL path, must different with Phase 1 URL path")
	listen    = flag.String("listen", ":10080", "Listen Address")
)

func registerProfileHandler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		file, handler, err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer file.Close()

		//fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile("./static/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer f.Close()

		io.Copy(f, file)
	} else {
		fmt.Println("Unknown HTTP " + r.Method + "  Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	flag.Parse()
	if *phase1Url == *phase2Url {
		flag.PrintDefaults()
		os.Exit(1)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc(*phase1Url, ht.Phase1Handler)
	mux.HandleFunc(*phase2Url, ht.Phase2Handler)

	mux.HandleFunc("/upload", upload)

	// Register pprof handlers
	registerProfileHandler(mux)

	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(*listen, mux)
}
