package main

import "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/file"

const (
	macSplit     = "/"
	windowsSplit = "\\"
)

func main() {
	//src := "C:\\Users\\zheng\\Desktop\\douyin\\src"
	dst := "C:\\Users\\zheng\\Desktop\\douyin\\dst"

	err := file.Remove(dst)
	if err != nil {
		println("执行错误：", err.Error())
	} else {
		println("执行成功")
	}

}
