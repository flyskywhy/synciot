package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	Version = "unknown-dev"
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
}
