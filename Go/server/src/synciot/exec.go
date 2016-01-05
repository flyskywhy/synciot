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
	stdoutMut = &sync.Mutex{}
)

type CmdServer struct {
	exit             chan error
	cmd              *exec.Cmd
	stdoutFirstLines []string // The first 10 lines of stdout
	stdoutLastLines  []string // The last 50 lines of stdout
}

func newCmdServer(workDir, bin string, arg ...string) *CmdServer {
	svr := &CmdServer{}

	svr.exit = make(chan error)

	svr.cmd = exec.Command(bin, arg...)
	svr.cmd.Dir = workDir

	return svr
}

func (s *CmdServer) Serve() {
	var dst io.Writer = os.Stdout

	stderr, err := s.cmd.StderrPipe()
	if err != nil {
		fmt.Println("stderr:", err)
	}

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		fmt.Println("stdout:", err)
	}

	fmt.Println("Starting", s.cmd.Args)
	err = s.cmd.Start()
	if err != nil {
		fmt.Println(err)
	}

	stdoutMut.Lock()
	s.stdoutFirstLines = make([]string, 0, 10)
	s.stdoutLastLines = make([]string, 0, 50)
	stdoutMut.Unlock()

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		s.copyStderr(stderr, dst)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		s.copyStdout(stdout, dst)
		wg.Done()
	}()

	go func() {
		wg.Wait()
		s.exit <- s.cmd.Wait()
	}()
}

func (s *CmdServer) Stop() {
	s.cmd.Process.Kill()
	<-s.exit
	fmt.Println(s.cmd.Args, "exited")
}

func (s *CmdServer) copyStderr(stderr io.Reader, dst io.Writer) {
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

func (s *CmdServer) copyStdout(stdout io.Reader, dst io.Writer) {
	br := bufio.NewReader(stdout)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return
		}

		stdoutMut.Lock()
		if len(s.stdoutFirstLines) < cap(s.stdoutFirstLines) {
			s.stdoutFirstLines = append(s.stdoutFirstLines, line)
		} else {
			if l := len(s.stdoutLastLines); l == cap(s.stdoutLastLines) {
				s.stdoutLastLines = s.stdoutLastLines[:l-1]
			}
			s.stdoutLastLines = append(s.stdoutLastLines, line)
		}
		stdoutMut.Unlock()

		dst.Write([]byte(line))
	}
}
