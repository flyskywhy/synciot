package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/thejerf/suture"
)

var (
	Version = "unknown-dev"
)

const (
	guiAssets  = "gui"
	guiAddress = "127.0.0.1:7777"

	CONFIG_JSON = "synciot.json"

	SYNCTHING_CONFIG_DIR = "config"
)

type FolderConfiguration struct {
	ID      string `json:"id"`
	RawPath string `json:"path"`
}

type Configuration struct {
	Folders []FolderConfiguration `json:"folders"`
}

var (
	binDir   string
	mainSvc  = suture.NewSimple("main")
	quitChan chan os.Signal
)

// Command line and environment options
var (
	showVersion bool
)

func init() {
	Version = "0.1"

	mainSvc.ServeBackground()
}

func main() {
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Parse()

	args := os.Args
	binPath, _ := filepath.Abs(args[0])
	binDir, _ = filepath.Split(binPath)
	fmt.Println("Starting", binPath)

	if showVersion {
		fmt.Println(Version)
		return
	}

	synciotMain()

	fmt.Println(binPath, "exited")
}

func synciotMain() {
	// Event subscription for the API; must start early to catch the early events.
	apiSub := NewBufferedSubscription(Default.Subscribe(AllEvents), 1000)

	setupGUI(apiSub)

	<-quitChan
	mainSvc.Stop()
}

func setupGUI(apiSub *BufferedSubscription) {
	assets := filepath.Join(binDir, guiAssets)
	config := filepath.Join(binDir, CONFIG_JSON)

	api, err := newAPISvc(assets, config, guiAddress, apiSub)
	if err != nil {
		fmt.Println("Cannot start GUI:", err)
	} else {
		fmt.Println("Starting GUI from", assets)
		fmt.Println("Synciot listening on", guiAddress)
	}
	mainSvc.Add(api)
}

// getFreePort returns a free TCP port fort listening on. The ports given are
// tried in succession and the first to succeed is returned. If none succeed,
// a random high port is returned.
func getFreePort(host string, ports ...int) (int, error) {
	for _, port := range ports {
		c, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
		if err == nil {
			c.Close()
			return port, nil
		}
	}

	c, err := net.Listen("tcp", host+":0")
	if err != nil {
		return 0, err
	}
	addr := c.Addr().(*net.TCPAddr)
	c.Close()
	return addr.Port, nil
}
