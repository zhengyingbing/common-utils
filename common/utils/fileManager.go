package utils

import (
	"errors"
	"os"
	"sync"
)

type FileMgr struct {
	fileHandler map[string]*os.File
	mutex       sync.Mutex
}

func CreateFileMgr() *FileMgr {
	return &FileMgr{
		fileHandler: make(map[string]*os.File),
	}
}

func (mgr *FileMgr) RemoveFile(path string) error {
	if Exist(path) {
		return os.Remove(path)
	}
	return nil
}

func (mgr *FileMgr) WriteFile(path string, content string) error {
	var file *os.File
	var err error
	if _, ok := mgr.fileHandler[path]; ok {
		file = mgr.fileHandler[path]
	} else {
		file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0766)
		if err != nil {
			os.Exit(-1)
		}

		mgr.mutex.Lock()
		mgr.fileHandler[path] = file
		mgr.mutex.Unlock()
	}
	_, err = file.WriteString(content)
	return err
}

func (mgr *FileMgr) End(path string) error {
	if file, ok := mgr.fileHandler[path]; ok {
		file.Close()
		delete(mgr.fileHandler, path)
		return nil
	}
	return errors.New("File not exist: " + path)
}
