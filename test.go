package main

import (
	"log"
	"path/filepath"
)

const (
	macSplit     = "/"
	windowsSplit = "\\"
)

func main() {
	//err := utils.ReplaceAllFiles("C:\\Users\\zheng\\Desktop\\douyin", "1111", "2222")
	str := "abc"
	log.Println(filepath.Join(str, "ac", "dd", "macos"))

}

type LoginCallback interface {
	OnSuccess(uid, token string)
	onFailed(err string)
}

type HandleLogin struct{}

func (h HandleLogin) OnSuccess(uid, token string) {

}
