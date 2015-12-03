package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type apiSvc struct {
	assetDir string
	listener net.Listener
	stop     chan struct{}
}

func newAPISvc(assets, address string) (*apiSvc, error) {
	svc := &apiSvc{
		assetDir: assets,
	}

	var err error
	svc.listener, err = net.Listen("tcp", address)
	return svc, err
}

func (s *apiSvc) Serve() {
	mux := http.NewServeMux()
	mux.Handle("/", embeddedStatic{
		assetDir: s.assetDir,
	})

	srv := http.Server{
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
	}

	err := srv.Serve(s.listener)

	select {
	case <-s.stop:
	case <-time.After(time.Second):
		fmt.Println("API:", err)
	}
}

func (s *apiSvc) Stop() {
}

type embeddedStatic struct {
	assetDir string
}

func (s embeddedStatic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path

	if file[0] == '/' {
		file = file[1:]
	}

	if len(file) == 0 {
		file = "index.html"
	}

	p := filepath.Join(s.assetDir, filepath.FromSlash(file))
	_, err := os.Stat(p)
	if err == nil {
		http.ServeFile(w, r, p)
		return
	} else {
		http.NotFound(w, r)
		return
	}
}
