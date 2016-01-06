package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/wuxicn/pipeline"
)

type apiSvc struct {
	cfg       *Configuration
	assetDir  string
	cmdServer map[string]*CmdServer
	cfgPath   string
	listener  net.Listener
	stop      chan struct{}
}

func newAPISvc(assets, config, address string) (*apiSvc, error) {
	svc := &apiSvc{
		assetDir:  assets,
		cmdServer: make(map[string]*CmdServer),
		cfgPath:   config,
	}

	var err error
	svc.listener, err = net.Listen("tcp", address)
	return svc, err
}

func (s *apiSvc) Serve() {
	s.stop = make(chan struct{})

	// The GET handlers
	getRestMux := http.NewServeMux()
	getRestMux.HandleFunc("/rest/stats/folder", s.getFolderStats)
	getRestMux.HandleFunc("/rest/system/config", s.getSystemConfig)
	getRestMux.HandleFunc("/rest/system/status", s.getSystemStatus)

	// The POST handlers
	postRestMux := http.NewServeMux()
	postRestMux.HandleFunc("/rest/system/config", s.postSystemConfig)
	postRestMux.HandleFunc("/rest/system/generate", s.postGenFolder)
	postRestMux.HandleFunc("/rest/system/start", s.postStartFolder)
	postRestMux.HandleFunc("/rest/system/stop", s.postStopFolder)

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

func (s *apiSvc) getSystemConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := ioutil.ReadFile(s.cfgPath)

	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(cfg)
}

func (s *apiSvc) fillCfgFromFile() {
	cfg_byte, err := ioutil.ReadFile(s.cfgPath)

	if err != nil {
		fmt.Println(err)
		return
	}

	json.Unmarshal(cfg_byte, &s.cfg)
}

func (s *apiSvc) getFolderStats(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	folder := qs.Get("folder")
	res := s.folderSummary(folder)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(res)
}

func (s *apiSvc) folderSummary(folder string) map[string]interface{} {
	var res = make(map[string]interface{})
	syncthingGuiPort := ""

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Folders != nil {
		for _, rf := range s.cfg.Folders {
			if rf.ID == folder {
				syncthingGuiPort = getSyncthingGuiPort(filepath.FromSlash(rf.RawPath + "/" + SYNCTHING_CONFIG_DIR + "/config.xml"))
				break
			}
		}
	}

	if syncthingGuiPort == "" {
		return res
	} else {
		res["syncthingGuiPort"] = syncthingGuiPort
	}

	req, err := http.NewRequest("GET", "http://127.0.0.1:"+syncthingGuiPort+"/rest/system/ping", nil)
	_, err = http.DefaultClient.Do(req)
	if err == nil {
		res["state"] = "running"
	} else {
		res["state"] = "stopped"
	}

	return res
}

func (s *apiSvc) postSystemConfig(w http.ResponseWriter, r *http.Request) {
	var cfg = make([]byte, r.ContentLength)
	r.Body.Read(cfg)

	err := ioutil.WriteFile(s.cfgPath, cfg, 0644)

	if err == nil {
		fmt.Println("Writed", s.cfgPath)
	}
}

func getSyncthingGuiPort(path string) string {
	port := ""

	_, err := os.Stat(path)
	if err != nil {
		return port
	}

	stdout, _, err := pipeline.Run(
		exec.Command("grep", "address", path),
		exec.Command("tail", "-1"),
		exec.Command("sed", "s/.*<address>.*://"),
		exec.Command("sed", "s/<\\/address>.*//"),
		exec.Command("tr", "-d", "\\\"[\\n][\\r]\\\""))

	if err != nil {
		e := err.(*pipeline.Error)
		fmt.Println("ERR:", e.Code, e.Err)
		return port
	}

	port = stdout.String()
	return port
}

func getSyncthingProtocolPort(path string) string {
	port := ""

	_, err := os.Stat(path)
	if err != nil {
		return port
	}

	stdout, _, err := pipeline.Run(
		exec.Command("grep", "listenAddress", path),
		exec.Command("sed", "s/.*<listenAddress>.*://"),
		exec.Command("sed", "s/<\\/listenAddress>.*//"),
		exec.Command("tr", "-d", "\\\"[\\n][\\r]\\\""))

	if err != nil {
		e := err.(*pipeline.Error)
		fmt.Println("ERR:", e.Code, e.Err)
		return port
	}

	port = stdout.String()
	return port
}

func setSyncthingGuiPort(path string, port string) {
	_, err := os.Stat(path)
	if err != nil {
		return
	}

	oldPort := getSyncthingGuiPort(path)

	_, _, err = pipeline.Run(
		exec.Command("sed", "-i", "s/:"+oldPort+"<\\/address>/:"+port+"<\\/address>/", path))

	if err != nil {
		e := err.(*pipeline.Error)
		fmt.Println("ERR:", e.Code, e.Err)
		return
	}
}

func setSyncthingProtocolPort(path string, port string) {
	_, err := os.Stat(path)
	if err != nil {
		return
	}

	oldPort := getSyncthingProtocolPort(path)

	_, _, err = pipeline.Run(
		exec.Command("sed", "-i", "s/:"+oldPort+"<\\/listenAddress>/:"+port+"<\\/listenAddress>/", path))

	if err != nil {
		e := err.(*pipeline.Error)
		fmt.Println("ERR:", e.Code, e.Err)
		return
	}
}

func setSyncthingFolderConnector(synciotDir string) {
	xmlDir := filepath.FromSlash(synciotDir + "/" + SYNCTHING_CONFIG_DIR)
	xmlPath := filepath.FromSlash(xmlDir + "/config.xml")

	_, err := os.Stat(xmlPath)
	if err != nil {
		return
	}

	buf, _ := ioutil.ReadFile(xmlPath)
	reg := regexp.MustCompile("id=\"default\" path=\".*\" ro=")
	buf = reg.ReplaceAll(buf, []byte("id=\"connector\" path=\""+synciotDir+string(filepath.Separator)+"connector\" ro="))
	ioutil.WriteFile(xmlPath, buf, 0644)
}

func (s *apiSvc) fromAllConfigXml(get func(string) string) []string {
	var values []string
	var value string

	s.fillCfgFromFile()

	if s.cfg == nil || s.cfg.Folders == nil {
		return values
	}

	for _, rf := range s.cfg.Folders {
		value = get(filepath.FromSlash(rf.RawPath + "/" + SYNCTHING_CONFIG_DIR + "/config.xml"))
		if value != "" {
			values = append(values, value)
		}
	}

	sort.Strings(values)
	return values
}

func getIncreasedPort(ports []string, host, defaultPort string) int {
	var port int

	if len(ports) == 0 {
		port, _ = strconv.Atoi(defaultPort)
		return port
	} else {
		port, _ = strconv.Atoi(ports[len(ports)-1])
		port++
		for {
			port_tmp, _ := getFreePort(host, port)

			if port_tmp != port {
				port++
			} else {
				return port
			}
		}
	}
}

func (s *apiSvc) postGenFolder(w http.ResponseWriter, r *http.Request) {
	guiPorts := s.fromAllConfigXml(getSyncthingGuiPort)
	guiPort := strconv.Itoa(getIncreasedPort(guiPorts, "127.0.0.1", "8384"))

	protocolPorts := s.fromAllConfigXml(getSyncthingProtocolPort)
	protocolPort := strconv.Itoa(getIncreasedPort(protocolPorts, "0.0.0.0", "22000"))

	qs := r.URL.Query()
	synciotDir := qs.Get("path")
	xmlDir := filepath.FromSlash(synciotDir + "/" + SYNCTHING_CONFIG_DIR)
	os.MkdirAll(xmlDir, 0775)

	cmd := exec.Command(filepath.Join(binDir, "syncthing"), "-generate="+xmlDir)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))

	xmlPath := filepath.FromSlash(xmlDir + "/config.xml")
	setSyncthingGuiPort(xmlPath, guiPort)
	setSyncthingProtocolPort(xmlPath, protocolPort)
	os.MkdirAll(filepath.FromSlash(synciotDir+"/connector"), 0775)
	setSyncthingFolderConnector(filepath.FromSlash(synciotDir))
}

func (s *apiSvc) postStartFolder(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	folder := qs.Get("folder")

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Folders != nil {
		for _, rf := range s.cfg.Folders {
			if rf.ID == folder {
				xmlDir := filepath.FromSlash(rf.RawPath + "/" + SYNCTHING_CONFIG_DIR)
				xmlPath := filepath.FromSlash(xmlDir + "/config.xml")
				_, err := os.Stat(xmlPath)
				if err == nil {
					port := getSyncthingGuiPort(xmlPath)

					cmd := newCmdServer(binDir, filepath.Join(binDir, "syncthing"), "-no-browser", "-no-restart", "-gui-address=127.0.0.1:"+port, "-home="+xmlDir)
					s.cmdServer[rf.ID] = cmd
					cmd.Serve()

					return
				} else {
					fmt.Println(err)
					http.Error(w, err.Error(), 500)
					return
				}

			}
		}
	}
}

func (s *apiSvc) postStopFolder(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	folder := qs.Get("folder")

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Folders != nil {
		for _, rf := range s.cfg.Folders {
			if rf.ID == folder {
				cmd := s.cmdServer[rf.ID]
				if cmd != nil {
					cmd.Stop()
					return
				} else {
					fmt.Println("Warning: No cmdServer with folder", rf.ID)
					return
				}

			}
		}
	}
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
