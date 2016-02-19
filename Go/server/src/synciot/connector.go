package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The connector service runs a loop for connecting server and client from file ${Synciot}/connector/${ClientId}.
type connectorSvc struct {
	connectorDir string
	syncDir      string
	xmlPath      string
	stop         chan struct{}
}

func newConnectorSvc(path string) *connectorSvc {
	return &connectorSvc{
		connectorDir: filepath.FromSlash(path + "/" + CONNECTOR_DIR),
		syncDir:      filepath.FromSlash(path + "/" + SYNC_DIR),
		xmlPath:      filepath.FromSlash(path + "/" + SYNCTHING_CONFIG_DIR + "/config.xml"),
	}
}

func (s *connectorSvc) Serve() {
	s.stop = make(chan struct{})

	time.Sleep(10 * time.Second) // Waiting for started syncthing

	for {
		dir, err := ioutil.ReadDir(s.connectorDir)

		if len(dir) > 0 {
			if err == nil {
				for _, fi := range dir {
					id := fi.Name()
					if len(id) == 63 { // the length of syncthing Device ID is 63
						if setSyncthingFolderDevice(s.syncDir, s.xmlPath, id) == nil {
							os.Remove(filepath.FromSlash(s.connectorDir + "/" + id))

							fmt.Println("Client device", id, "added successfully")
						} else {
							fmt.Println("Warning: Client device", id, "added failed")
						}
					}
				}
			}
		}

		select {
		case <-s.stop:
			return
		case <-time.After(10 * time.Second):
		}
	}
}

func (s *connectorSvc) Stop() {
	close(s.stop)
}

func setSyncthingFolderDevice(syncDir, xmlPath, id string) error {
	_, err := os.Stat(xmlPath)
	if err != nil {
		return err
	}
	buf, _ := ioutil.ReadFile(xmlPath)
	xml := string(buf)

	shortId := getSyncthingDeviceIdShort(id)
	folderPath := filepath.FromSlash(syncDir + "/" + shortId + "-temp")
	os.MkdirAll(folderPath, 0775)
	os.Create(filepath.FromSlash(folderPath + "/.stfolder"))

	folderDevice := strings.Join(CLIENT_EXTRA_FOLDER_DEVICE, "")
	folderDevice = strings.Replace(folderDevice, "FOLDER_ID", shortId+"-Temp", -1)
	folderDevice = strings.Replace(folderDevice, "FOLDER_PATH", folderPath, -1)
	folderDevice = strings.Replace(folderDevice, "SERVER_DEVICE_ID", getSyncthingMyId(xmlPath), -1)
	folderDevice = strings.Replace(folderDevice, "CLIENT_DEVICE_ID", id, -1)

	xml = StringsInsert(xml, "    </folder>\n", folderDevice)

	device := strings.Join(CLIENT_DEVICE, "")
	device = strings.Replace(device, "CLIENT_DEVICE_ID", id, -1)
	device = strings.Replace(device, "CLIENT_DEVICE_NAME", shortId, -1)

	xml = StringsInsert(xml, "    </device>\n", device)

	ioutil.WriteFile(xmlPath, []byte(xml), 0644)

	return err
}