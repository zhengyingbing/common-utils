package main

import (
	"fmt"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging"
)

const (
	macSplit     = "/"
	windowsSplit = "\\"
)

func main() {
	data := []int{0, 1, 1, 2}
	a := packaging.Duplicate(data)
	fmt.Println(a)
}

type LoginCallback interface {
	OnSuccess(uid, token string)
	onFailed(err string)
}

type HandleLogin struct{}

func (h HandleLogin) OnSuccess(uid, token string) {

}
