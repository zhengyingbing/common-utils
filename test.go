package main

import (
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
)

const (
	macSplit     = "/"
	windowsSplit = "\\"
)

func main() {
	//data := []int{0, 1, 1, 2}
	//os.MkdirAll("C://apktool//home//3015_10302//gameDir//gen//bin", os.ModePerm)
	a := "C:\\apktool\\home\\3015_10302\\douyinDir\\res\\drawable-v23"
	b := "C:\\apktool\\home\\3015_10302\\gameDir\\res\\drawable-v23"
	//p2 := "C://apktool//home//3015_10302//gameDir//gen//bin//2.txt"
	utils.Move(a, b, false)

	//f2, err := os.Open(p2)
	//if err != nil {
	//	println("open:", err.Error())
	//}
	//_, err = io.Copy(f1, f2)
	//if err != nil {
	//	println("copy:", err.Error())
	//}

	//data := make(map[string]any)
	//data["a"] = 1
	//data["b"] = "aaa"
	//data["c"] = 1
	//collect := append(data["a"].([]interface{}), data["b"].([]interface{}))
	//a := utils.Duplicate(collect)
	//fmt.Println(a)
}

type LoginCallback interface {
	OnSuccess(uid, token string)
	onFailed(err string)
}

type HandleLogin struct{}

func (h HandleLogin) OnSuccess(uid, token string) {

}
