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
	IO_DIR               = "io"
	IN_DIR               = "in"
	OUT_DIR              = "out"
	SYNC_DIR             = "sync"
	CONNECTOR_DIR        = "connector"
)

type ServerConfiguration struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type Configuration struct {
	Servers []ServerConfiguration `json:"servers"`
}

var (
	binDir   string
	mainSvc  = suture.NewSimple("main")
	quitChan chan os.Signal

	CLIENT_FOLDER_DEVICE = []string{
		"    <folder id=\"FOLDER_ID\" path=\"FOLDER_PATH/\" ro=\"false\" rescanIntervalS=\"60\" ignorePerms=\"false\" autoNormalize=\"true\">\n",
		"        <device id=\"SERVER_DEVICE_ID\"></device>\n",
		"        <device id=\"CLIENT_DEVICE_ID\"></device>\n",
		"        <minDiskFreePct>1</minDiskFreePct>\n",
		"        <versioning></versioning>\n",
		"        <copiers>0</copiers>\n",
		"        <pullers>0</pullers>\n",
		"        <hashers>0</hashers>\n",
		"        <order>random</order>\n",
		"        <ignoreDelete>false</ignoreDelete>\n",
		"        <scanProgressIntervalS>0</scanProgressIntervalS>\n",
		"        <pullerSleepS>0</pullerSleepS>\n",
		"        <pullerPauseS>0</pullerPauseS>\n",
		"        <maxConflicts>-1</maxConflicts>\n",
		"        <disableSparseFiles>false</disableSparseFiles>\n",
		"    </folder>\n",
	}

	CLIENT_DEVICE = []string{
		"    <device id=\"CLIENT_DEVICE_ID\" name=\"CLIENT_DEVICE_NAME\" compression=\"metadata\" introducer=\"false\">\n",
		"        <address>dynamic</address>\n",
		"    </device>\n",
	}
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
	setupGUI()

	<-quitChan
	mainSvc.Stop()
}

func setupGUI() {
	assets := filepath.Join(binDir, guiAssets)
	config := filepath.Join(binDir, CONFIG_JSON)

	api, err := newAPISvc(assets, config, guiAddress)
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
