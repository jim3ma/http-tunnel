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
	authUser  = flag.String("auth-username", "default", "Auth Username")
	authPass  = flag.String("auth-password", "default-pass", "Auth Password")
)

func registerProfileHandler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", auth(pprof.Index))
	//mux.HandleFunc("/debug/pprof/cmdline", auth(pprof.Cmdline))
	mux.HandleFunc("/debug/pprof/profile", auth(pprof.Profile))
	mux.HandleFunc("/debug/pprof/symbol", auth(pprof.Symbol))
	mux.HandleFunc("/debug/pprof/trace", auth(pprof.Trace))
}

func auth(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()
		if !check(user, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="please enter credential"`)
			http.Error(w, "Unauthorized.", 401)
			return
		}
		fn(w, r)
	}
}

func check(user, pass string) bool {
	return user == *authUser && pass == *authPass
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
	mux.HandleFunc(*phase1Url, auth(ht.Phase1Handler))
	mux.HandleFunc(*phase2Url, auth(ht.Phase2Handler))

	mux.HandleFunc("/upload", auth(upload))

	// Register pprof handlers
	registerProfileHandler(mux)

	fs := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(*listen, mux)
}
