package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/evalgo/evos"
)

// The out service runs a loop for discovery of ${Synciot}/sync/${Client}-temp/${DateTime}/out.*.synciot
// and move to ${Synciot}/io/user${userIdNum}/out/${Client}-temp/${DateTime}/.
type outSvc struct {
	syncDir string
	ioDir   string
	stop    chan struct{}
}

func newOutSvc(path string) *outSvc {
	return &outSvc{
		syncDir: filepath.FromSlash(path + "/" + SYNC_DIR),
		ioDir:   filepath.FromSlash(path + "/" + IO_DIR),
	}
}

func (s *outSvc) Serve() {
	s.stop = make(chan struct{})

	for {
		outFiles := s.listenForMove()
		if len(outFiles) > 0 {
			for _, file := range outFiles {
				name := filepath.Base(file)
				i := strings.IndexRune(name, '.')
				userIdNum := name[len(OUT_DIR):i]
				syncDateDir := filepath.Dir(file)
				dateTimeDirBase := filepath.Base(syncDateDir)
				clientTempDir := filepath.Dir(syncDateDir)
				clientTempDirBase := filepath.Base(clientTempDir)
				ioDateDir := filepath.FromSlash(s.ioDir + "/user" + userIdNum + "/" + OUT_DIR + "/" + clientTempDirBase + "/" + dateTimeDirBase)

				evos.MoveFolder(syncDateDir, ioDateDir)
			}
		}

		select {
		case <-s.stop:
			return
		case <-time.After(10 * time.Second):
		}
	}
}

func (s *outSvc) Stop() {
	close(s.stop)
}

func (s *outSvc) listenForMove() []string {
	var outFiles []string
	prefix := strings.ToUpper("out")
	suffix := strings.ToUpper(".synciot")

	filepath.Walk(s.syncDir, func(filename string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}

		name := fi.Name()
		if strings.HasPrefix(strings.ToUpper(name), prefix) {
			if strings.HasSuffix(strings.ToUpper(name), suffix) {
				i := strings.IndexRune(name, '.')
				count, _ := strconv.Atoi(name[i+1 : len(name)-len(suffix)])
				folder := filepath.Dir(filename)
				if count == CountFiles(folder)-1 { // out*.synciot itself was not counted by client, so `-1` here
					outFiles = append(outFiles, filename)
				}
			}
		}

		return nil
	})

	return outFiles
}
