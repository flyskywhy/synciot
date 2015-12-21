package main

import (
	"flag"
	"fmt"
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
)

type FolderConfiguration struct {
	ID      string `json:"id"`
	RawPath string `json:"path"`
}

type Configuration struct {
	Folders []FolderConfiguration `json:"folders"`
}

type Model struct {
	State string `json:"state"`
}

var (
	binDir   string
	quitChan chan os.Signal
)

// Command line and environment options
var (
	showVersion bool
)

func init() {
	Version = "0.1"
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
	mainSvc := suture.NewSimple("main")
	mainSvc.ServeBackground()

	setupGUI(mainSvc)

	runCmd(binDir, "echo", "Hello")

	<-quitChan
	mainSvc.Stop()
}

func setupGUI(mainSvc *suture.Supervisor) {
	assets := filepath.Join(binDir, guiAssets)
	config := filepath.Join(binDir, CONFIG_JSON)

	api, err := newAPISvc(assets, config, guiAddress)
	if err != nil {
		fmt.Println("Cannot start GUI:", err)
	} else {
		fmt.Println("Starting GUI from", assets)
		fmt.Println("API listening on", guiAddress)
	}
	mainSvc.Add(api)
}
