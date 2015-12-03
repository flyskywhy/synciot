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
)

var (
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
	fmt.Println(args)

	if showVersion {
		fmt.Println(Version)
		return
	}

	synciotMain()
}

func synciotMain() {
	mainSvc := suture.NewSimple("main")
	mainSvc.ServeBackground()

	setupGUI(mainSvc)

	<-quitChan
	mainSvc.Stop()
}

func setupGUI(mainSvc *suture.Supervisor) {
	assets := filepath.Join(filepath.Dir(os.Args[0]), guiAssets)
	api, err := newAPISvc(assets, guiAddress)
	if err != nil {
		fmt.Println("Cannot start GUI:", err)
	} else {
		fmt.Println("Starting GUI from", assets)
		fmt.Println("API listening on", guiAddress)
	}
	mainSvc.Add(api)
}
