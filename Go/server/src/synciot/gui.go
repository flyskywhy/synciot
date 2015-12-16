package main

import (
	"encoding/json"
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
	s.stop = make(chan struct{})

	// The GET handlers
	getRestMux := http.NewServeMux()
	getRestMux.HandleFunc("/rest/system/status", s.getSystemStatus)

	// The POST handlers
	postRestMux := http.NewServeMux()
	postRestMux.HandleFunc("/rest/system/config", s.postSystemConfig)

	// A handler that splits requests between the two above and disables
	// caching
	restMux := noCacheMiddleware(getPostHandler(getRestMux, postRestMux))

	// The main routing handler
	mux := http.NewServeMux()
	mux.Handle("/rest/", restMux)

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

func getPostHandler(get, post http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			get.ServeHTTP(w, r)
		case "POST":
			post.ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func noCacheMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=0, no-cache, no-store")
		w.Header().Set("Expires", time.Now().UTC().Format(http.TimeFormat))
		w.Header().Set("Pragma", "no-cache")
		h.ServeHTTP(w, r)
	})
}

type FolderConfiguration struct {
	ID      string `json:"id"`
	RawPath string `json:"path"`
}

type Configuration struct {
	Folders []FolderConfiguration `json:"folders"`
}

func (s *apiSvc) postSystemConfig(w http.ResponseWriter, r *http.Request) {
	var cfg Configuration

	err := json.NewDecoder(r.Body).Decode(&cfg)

	if err != nil {
		fmt.Println("decoding posted config:", err)
		http.Error(w, err.Error(), 500)
		return
	}

	for key, value := range cfg.Folders {
		fmt.Println("Key:", key, "Value:", value)
		fmt.Println("Id:", value.ID, "Path:", value.RawPath)
	}

	fmt.Println(cfg.Folders)
}

func (s *apiSvc) getSystemStatus(w http.ResponseWriter, r *http.Request) {
	tilde, _ := ExpandTilde("~")
	res := make(map[string]interface{})
	res["tilde"] = tilde
	res["pathSeparator"] = string(filepath.Separator)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(res)
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
