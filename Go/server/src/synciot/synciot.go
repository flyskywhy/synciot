package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/thejerf/suture"
)

var (
	Version = "unknown-dev"
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
	svc := &apiSvc{}
	mainSvc.Add(svc)
}

type apiSvc struct {
}

func (s *apiSvc) Serve() {
	for {
		time.Sleep(time.Second)
		fmt.Println("Hello")
	}
}

func (s *apiSvc) Stop() {
}
