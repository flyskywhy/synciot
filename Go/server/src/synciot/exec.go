package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

var (
	stdoutFirstLines []string // The first 10 lines of stdout
	stdoutLastLines  []string // The last 50 lines of stdout
	stdoutMut        = &sync.Mutex{}
)

func runCmd(workDir, bin string, arg ...string) {
	var dst io.Writer = os.Stdout

	cmd := exec.Command(bin, arg...)
	cmd.Dir = workDir

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("stderr:", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("stdout:", err)
	}

	fmt.Println("Starting", cmd.Args)
	err = cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	stdoutMut.Lock()
	stdoutFirstLines = make([]string, 0, 10)
	stdoutLastLines = make([]string, 0, 50)
	stdoutMut.Unlock()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		copyStderr(stderr, dst)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		copyStdout(stdout, dst)
		wg.Done()
	}()

	wg.Wait()
	cmd.Wait()
	fmt.Println(cmd.Args, "exited")
}

func copyStderr(stderr io.Reader, dst io.Writer) {
	br := bufio.NewReader(stderr)

	var panicFd *os.File
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return
		}

		if panicFd == nil {
			dst.Write([]byte(line))

			panicFd.WriteString("Panic at " + time.Now().Format(time.RFC3339) + "\n")
		}

		if panicFd != nil {
			panicFd.WriteString(line)
		}
	}
}

func copyStdout(stdout io.Reader, dst io.Writer) {
	br := bufio.NewReader(stdout)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return
		}

		stdoutMut.Lock()
		if len(stdoutFirstLines) < cap(stdoutFirstLines) {
			stdoutFirstLines = append(stdoutFirstLines, line)
		} else {
			if l := len(stdoutLastLines); l == cap(stdoutLastLines) {
				stdoutLastLines = stdoutLastLines[:l-1]
			}
			stdoutLastLines = append(stdoutLastLines, line)
		}
		stdoutMut.Unlock()

		dst.Write([]byte(line))
	}
}
