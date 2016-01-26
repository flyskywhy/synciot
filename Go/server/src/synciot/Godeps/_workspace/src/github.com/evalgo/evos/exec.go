// Copyright 2015 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evos

import (
	"bytes"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

// this function is a wrapper for the os/exec::Command
// it invokes the -debug flag for the eve commands and handles
// by default 2 streams ERR and OUT this is sometimes very usefull
// for example git commands are tricky because the seems to use different
// streams in a strange wrong manner
func ECommand(name string, arg ...string) ([]byte, error) {
	ecmd := exec.Command(name, arg...)
	if DEBUG {
		// todo:
		// check if this is a golang tools which is used in this case
		// do not add -debug flag to the command
		if strings.Contains(filepath.Base(name), "eve") {
			ecmd.Args = append(ecmd.Args, "-debug")
		}
		log.Println("ECommand running command:")
		log.Println(name)
		log.Println("with params:")
		log.Println(ecmd.Args)
	}
	stdout := bytes.NewBuffer(nil)
	ecmd.Stdout = stdout
	stderr := bytes.NewBuffer(nil)
	ecmd.Stderr = stderr
	err := ecmd.Run()
	if DEBUG {
		log.Println("ECommand STDERR", stderr.String())
		log.Println("ECommand STDOUT", stdout.String())
		log.Println("ECommand error:", err)
	}
	if err != nil {
		return nil, err
	}
	return stdout.Bytes(), nil
}
