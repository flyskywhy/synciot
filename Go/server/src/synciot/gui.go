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

	"github.com/evalgo/evos"
	"github.com/thejerf/suture"
)

type apiSvc struct {
	cfg            *Configuration
	assetDir       string
	cmdServer      map[string]*CmdServer
	outSvcId       map[string]suture.ServiceToken
	connectorSvcId map[string]suture.ServiceToken
	cfgPath        string
	listener       net.Listener
	stop           chan struct{}
}

func newAPISvc(assets, config, address string) (*apiSvc, error) {
	svc := &apiSvc{
		assetDir:       assets,
		cmdServer:      make(map[string]*CmdServer),
		outSvcId:       make(map[string]suture.ServiceToken),
		connectorSvcId: make(map[string]suture.ServiceToken),
		cfgPath:        config,
	}

	var err error
	svc.listener, err = net.Listen("tcp", address)
	return svc, err
}

func (s *apiSvc) Serve() {
	s.stop = make(chan struct{})

	// The GET handlers
	getRestMux := http.NewServeMux()
	getRestMux.HandleFunc("/rest/server/config", s.getServerConfig)
	getRestMux.HandleFunc("/rest/server/status", s.getServerStatus)
	getRestMux.HandleFunc("/rest/client/config", s.getClientConfig)
	getRestMux.HandleFunc("/rest/client/status", s.getClientStatus)
	getRestMux.HandleFunc("/rest/system/status", s.getSystemStatus)

	// The POST handlers
	postRestMux := http.NewServeMux()
	postRestMux.HandleFunc("/rest/server/config", s.postServerConfig)
	postRestMux.HandleFunc("/rest/server/generate", s.postGenServer)
	postRestMux.HandleFunc("/rest/server/start", s.postStartServer)
	postRestMux.HandleFunc("/rest/server/stop", s.postStopServer)
	postRestMux.HandleFunc("/rest/client/start", s.postStartClient)
	postRestMux.HandleFunc("/rest/client/stop", s.postStopClient)

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

type SyncthingSystemStatusConfiguration struct {
	MyID string `json:"myID"`
}

func getSyncthingMyId(xmlPath string) string {
	myID := ""
	url := "http://127.0.0.1:" + getSyncthingGuiPort(xmlPath) + "/rest/system/status"

	req, err := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		//fmt.Println("http.Do failed,[err=%s][url=%s]", err, url)
		return myID
	}
	defer resp.Body.Close()

	var cfg SyncthingSystemStatusConfiguration
	err = json.NewDecoder(resp.Body).Decode(&cfg)
	if err != nil {
		//fmt.Println("decoding posted config:", err)
		return myID
	}

	myID = cfg.MyID

	return myID
}

type ClientConfiguration struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserConfiguration struct {
	Clients []ClientConfiguration `json:"clients"`
}

func getSyncthingRemoteDevices(xmlPath string) []ClientConfiguration {
	var client ClientConfiguration
	var clients []ClientConfiguration
	id := ""
	myID := getSyncthingMyId(xmlPath)

	if myID == "" {
		return clients
	}

	buf, _ := ioutil.ReadFile(xmlPath)
	reg := regexp.MustCompile(".*<device id=\".*\" name=.*")
	devices := reg.FindAllString(string(buf), -1)
	for _, rf := range devices {
		reg = regexp.MustCompile(".*<device id=\"|\" name=.*")
		id = reg.ReplaceAllString(rf, "")
		if id != myID {
			client.ID = id
			reg = regexp.MustCompile(".*\" name=\"|\" compression=.*")
			client.Name = reg.ReplaceAllString(rf, "")
			clients = append(clients, client)
		}
	}

	return clients
}

func getSyncthingDeviceIdShort(id string) string {
	reg := regexp.MustCompile("-.*")
	return reg.ReplaceAllString(id, "")
}

func getClients(serverPath string) []ClientConfiguration {
	var client ClientConfiguration
	var clients []ClientConfiguration

	dir, err := ioutil.ReadDir(filepath.FromSlash(serverPath + "/" + SYNC_DIR))
	if err == nil {
		for _, fi := range dir {
			if fi.IsDir() {
				client.ID = fi.Name()
				client.Name = client.ID
				clients = append(clients, client)
			}
		}
	} else {
		fmt.Println("Error: ioutil.ReadDir() failed in getClients()")
	}

	return clients
}

func (s *apiSvc) getClientConfig(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	server := qs.Get("server")

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Servers != nil {
		for _, rf := range s.cfg.Servers {
			if rf.ID == server {
				var cfg UserConfiguration

				// Replace getClients() with inline code for more efficient
				//cfg.Clients = getClients(rf.Path)

				var client ClientConfiguration

				dir, err := ioutil.ReadDir(filepath.FromSlash(rf.Path + "/" + SYNC_DIR))
				if err == nil {
					for _, fi := range dir {
						if fi.IsDir() {
							client.ID = fi.Name()
							client.Name = client.ID
							cfg.Clients = append(cfg.Clients, client)
						}
					}
				} else {
					fmt.Println("Error: ioutil.ReadDir() failed in getClientConfig()")
				}

				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				json.NewEncoder(w).Encode(cfg)

				return
			}
		}
	}
}

func (s *apiSvc) getClientStatus(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	res := s.clientSummary(qs.Get("serverId"), qs.Get("clientId"), qs.Get("userIdNum"))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(res)

}

func (s *apiSvc) clientSummary(serverId, clientId, userIdNum string) map[string]interface{} {
	var res = make(map[string]interface{})

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Servers != nil {
		for _, rf := range s.cfg.Servers {
			if rf.ID == serverId {
				syncInDir := filepath.FromSlash(rf.Path + "/" + SYNC_DIR + "/" + clientId + "/" + IN_DIR)

				_, err := os.Stat(syncInDir)
				if err == nil {
					res["state"] = "syncing"
				} else {
					res["state"] = "idle"
				}

				outDir := filepath.FromSlash(rf.Path + "/" + IO_DIR + "/user" + userIdNum + "/" + OUT_DIR + "/" + clientId)
				res["out"] = CountDirs(outDir)

				return res
			}
		}
	}

	return res
}

func (s *apiSvc) getServerConfig(w http.ResponseWriter, r *http.Request) {
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

func (s *apiSvc) getServerStatus(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	server := qs.Get("server")
	res := s.serverSummary(server)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(res)
}

func (s *apiSvc) serverSummary(server string) map[string]interface{} {
	var res = make(map[string]interface{})
	syncthingGuiPort := ""

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Servers != nil {
		for _, rf := range s.cfg.Servers {
			if rf.ID == server {
				syncthingGuiPort = getSyncthingGuiPort(filepath.FromSlash(rf.Path + "/" + SYNCTHING_CONFIG_DIR + "/config.xml"))
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
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		res["state"] = "running"
		defer resp.Body.Close()
	} else {
		res["state"] = "stopped"
	}

	return res
}

func (s *apiSvc) postServerConfig(w http.ResponseWriter, r *http.Request) {
	var cfg = make([]byte, r.ContentLength)
	r.Body.Read(cfg)

	err := ioutil.WriteFile(s.cfgPath, cfg, 0644)

	if err == nil {
		fmt.Println("Writed", s.cfgPath)
	}
}

func getSyncthingGuiPort(xmlPath string) string {
	port := ""

	_, err := os.Stat(xmlPath)
	if err != nil {
		return port
	}

	buf, _ := ioutil.ReadFile(xmlPath)
	reg := regexp.MustCompile("<gui enabled.*\\s.*<address>.*")
	addr := string(reg.Find(buf))
	reg = regexp.MustCompile("<gui enabled.*\\s.*<address>.*:|</address>.*")
	port = reg.ReplaceAllString(addr, "")

	return port
}

func getSyncthingProtocolPort(xmlPath string) string {
	port := ""

	_, err := os.Stat(xmlPath)
	if err != nil {
		return port
	}

	buf, _ := ioutil.ReadFile(xmlPath)
	reg := regexp.MustCompile(".*<listenAddress>.*")
	addr := string(reg.Find(buf))
	reg = regexp.MustCompile(".*<listenAddress>.*://.*:|</listenAddress>.*")
	port = reg.ReplaceAllString(addr, "")

	return port
}

func setSyncthingGuiPort(xmlPath string, port string) {
	_, err := os.Stat(xmlPath)
	if err != nil {
		return
	}

	oldPort := getSyncthingGuiPort(xmlPath)

	buf, _ := ioutil.ReadFile(xmlPath)
	reg := regexp.MustCompile(":" + oldPort + "</address>")
	buf = reg.ReplaceAll(buf, []byte(":"+port+"</address>"))
	ioutil.WriteFile(xmlPath, buf, 0644)
}

func setSyncthingProtocolPort(xmlPath string, port string) {
	_, err := os.Stat(xmlPath)
	if err != nil {
		return
	}

	oldPort := getSyncthingProtocolPort(xmlPath)

	buf, _ := ioutil.ReadFile(xmlPath)
	reg := regexp.MustCompile(":" + oldPort + "</listenAddress>")
	buf = reg.ReplaceAll(buf, []byte(":"+port+"</listenAddress>"))
	ioutil.WriteFile(xmlPath, buf, 0644)
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
	buf = reg.ReplaceAll(buf, []byte("id=\"connector\" path=\""+synciotDir+string(filepath.Separator)+"connector/\" ro="))
	ioutil.WriteFile(xmlPath, buf, 0644)
}

func setSyncthingMisc(xmlPath string) {
	_, err := os.Stat(xmlPath)
	if err != nil {
		return
	}

	buf, _ := ioutil.ReadFile(xmlPath)
	reg := regexp.MustCompile("urAccepted>0")
	buf = reg.ReplaceAll(buf, []byte("urAccepted>-1"))
	reg = regexp.MustCompile("autoUpgradeIntervalH>12")
	buf = reg.ReplaceAll(buf, []byte("autoUpgradeIntervalH>0"))
	ioutil.WriteFile(xmlPath, buf, 0644)
}

func (s *apiSvc) fromAllConfigXml(get func(string) string) []string {
	var values []string
	var value string

	s.fillCfgFromFile()

	if s.cfg == nil || s.cfg.Servers == nil {
		return values
	}

	for _, rf := range s.cfg.Servers {
		value = get(filepath.FromSlash(rf.Path + "/" + SYNCTHING_CONFIG_DIR + "/config.xml"))
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

func genUserHtml(user string) {
	userHtml := filepath.FromSlash(binDir + "/gui/user-" + user + ".html")
	Copy(filepath.FromSlash(binDir+"/gui/user.html"), userHtml)
}

func (s *apiSvc) postGenServer(w http.ResponseWriter, r *http.Request) {
	guiPorts := s.fromAllConfigXml(getSyncthingGuiPort)
	guiPort := strconv.Itoa(getIncreasedPort(guiPorts, "127.0.0.1", "8384"))

	protocolPorts := s.fromAllConfigXml(getSyncthingProtocolPort)
	protocolPort := strconv.Itoa(getIncreasedPort(protocolPorts, "0.0.0.0", "22000"))

	qs := r.URL.Query()
	synciotDir := qs.Get("path")
	xmlDir := filepath.FromSlash(synciotDir + "/" + SYNCTHING_CONFIG_DIR)
	xmlPath := filepath.FromSlash(xmlDir + "/config.xml")
	_, err := os.Stat(xmlPath)
	if err != nil {
		os.MkdirAll(xmlDir, 0775)

		cmd := exec.Command(filepath.Join(binDir, "syncthing"), "-generate="+xmlDir)
		out, err := cmd.Output()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(out))

		setSyncthingGuiPort(xmlPath, guiPort)
		setSyncthingProtocolPort(xmlPath, protocolPort)
		os.MkdirAll(filepath.FromSlash(synciotDir+"/"+IO_DIR+"/user0/in"), 0775)
		os.MkdirAll(filepath.FromSlash(synciotDir+"/"+SYNC_DIR), 0775)
		os.MkdirAll(filepath.FromSlash(synciotDir+"/"+CONNECTOR_DIR), 0775)
		setSyncthingFolderConnector(filepath.FromSlash(synciotDir))
		setSyncthingMisc(xmlPath)
	} else {
		fmt.Println("Warning:", xmlPath, "aleady exist, if you want a new one, please remove config folder manually.")
	}

	genUserHtml(qs.Get("id"))
}

func (s *apiSvc) postStartServer(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	server := qs.Get("server")

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Servers != nil {
		for _, rf := range s.cfg.Servers {
			if rf.ID == server {
				xmlDir := filepath.FromSlash(rf.Path + "/" + SYNCTHING_CONFIG_DIR)
				xmlPath := filepath.FromSlash(xmlDir + "/config.xml")
				_, err := os.Stat(xmlPath)
				if err == nil {
					port := getSyncthingGuiPort(xmlPath)

					cmd := newCmdServer(binDir, filepath.Join(binDir, "syncthing"), "-no-browser", "-no-restart", "-gui-address=127.0.0.1:"+port, "-home="+xmlDir)
					s.cmdServer[rf.ID] = cmd
					cmd.Serve()

					s.outSvcId[rf.ID] = mainSvc.Add(newOutSvc(rf.Path))

					s.connectorSvcId[rf.ID] = mainSvc.Add(newConnectorSvc(rf.Path))

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

func (s *apiSvc) postStopServer(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	server := qs.Get("server")

	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Servers != nil {
		for _, rf := range s.cfg.Servers {
			if rf.ID == server {
				cmd := s.cmdServer[rf.ID]
				if cmd != nil {
					cmd.Stop()

					err := mainSvc.Remove(s.outSvcId[rf.ID])
					if err != nil {
						fmt.Println("Warning: Removing outSvc somehow failed with server", rf.ID)
					}

					err = mainSvc.Remove(s.connectorSvcId[rf.ID])
					if err != nil {
						fmt.Println("Warning: Removing connectorSvc somehow failed with server", rf.ID)
					}

					return
				} else {
					fmt.Println("Warning: No cmdServer with server", rf.ID)
					http.Error(w, "Warning: No cmdServer with server"+rf.ID, 500)
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

func (s *apiSvc) inClient(cmd, serverId, userIdNum string, result []byte) {
	s.fillCfgFromFile()
	if s.cfg != nil && s.cfg.Servers != nil {
		for _, rf := range s.cfg.Servers {
			if rf.ID == serverId {
				inDir := filepath.FromSlash(rf.Path + "/" + IO_DIR + "/user" + userIdNum + "/" + IN_DIR)
				os.MkdirAll(inDir, 0775)

				fc := CountFiles(inDir)
				os.Create(filepath.FromSlash(inDir + "/user" + userIdNum + "." + strconv.Itoa(fc) + "." + cmd + ".synciot"))

				if len(result) == 0 {
					// Replace getClients() with inline code for more efficient
					//for _, clientId := range getClients(rf.Path) {
					//	syncInDir := filepath.FromSlash(rf.Path + "/" + SYNC_DIR + "/" + clientId + "/" + IN_DIR)
					//	evos.CopyFolder(inDir, syncInDir)
					//}

					dir, err := ioutil.ReadDir(filepath.FromSlash(rf.Path + "/" + SYNC_DIR))
					if err == nil {
						for _, fi := range dir {
							if fi.IsDir() {
								syncInDir := filepath.FromSlash(rf.Path + "/" + SYNC_DIR + "/" + fi.Name() + "/" + IN_DIR)
								evos.CopyFolder(inDir, syncInDir)
							}
						}
					} else {
						fmt.Println("Error: ioutil.ReadDir() failed in inClient()")
					}
				} else {
					var clientIds []string
					json.Unmarshal(result, &clientIds)
					for _, clientId := range clientIds {
						syncInDir := filepath.FromSlash(rf.Path + "/" + SYNC_DIR + "/" + clientId + "/" + IN_DIR)
						evos.CopyFolder(inDir, syncInDir)
					}
				}

				os.RemoveAll(inDir)
				os.MkdirAll(inDir, 0775)
			}
		}
	}
}

func (s *apiSvc) postStartClient(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	serverId := qs.Get("serverId")
	userIdNum := qs.Get("userIdNum")
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	s.inClient("start", serverId, userIdNum, result)
}

func (s *apiSvc) postStopClient(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	serverId := qs.Get("serverId")
	userIdNum := qs.Get("userIdNum")
	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()

	s.inClient("stop", serverId, userIdNum, result)
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
