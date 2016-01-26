// Copyright 2015 EVALGO Community Developers.
// See the LICENSE file at the top-level directory of this distribution
// and at http://opensource.org/licenses/bsd-license.php

package evos

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	OsCopyDst = ""
)

func EveArch() string {
	return runtime.GOARCH
}

// checks if a file or folder exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// checks if the given path is folder or not
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), err
}

// copies a source to destination
func CopyFile(src, dst string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

// helper function for recursive copying
func WalkCopyFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		if DEBUG {
			log.Println("WalkCopyFunc copy file from", path, "to", OsCopyDst+string(os.PathSeparator)+path)
		}
		_, err = CopyFile(path, OsCopyDst+string(os.PathSeparator)+path)
		if err != nil {
			return err
		}
	} else {
		if DEBUG {
			log.Println("WalkCopyFunc create folder", OsCopyDst+string(os.PathSeparator)+path)
		}
		err = os.MkdirAll(OsCopyDst+string(os.PathSeparator)+path, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

// copies a folder from source to destination
func CopyFolder(src, dst string) error {
	OsCopyDst = dst
	err := filepath.Walk(src, WalkCopyFunc)
	if err != nil {
		return err
	}
	return nil
}

// moves a folder from a source to a destigation
func MoveFolder(src, dst string) error {
	err := CopyFolder(src, dst)
	if err != nil {
		return err
	}
	return os.RemoveAll(src)
}
