package main

import (
	"log"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/models"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/utils"
	"strconv"
)

const (
	product   = "3015"
	channelId = "10302"
	channel   = "douyin"
	game      = "aygd"
)

func main() {
	path, _ := os.Getwd()
	homePath := filepath.Join(path, "home")
	cfg := make(map[string]any)
	cfg["appName"] = channel + "Demo"
	cfg["targetSdkVersion"] = "30"
	cfg["dexMethodCounters"] = "60000"
	cfg["bundleId"] = "com.hoolai.qsmy.douyu"
	cfg["appId"] = "dypllxgp03osw"

	androidHome := filepath.Join(path, "resources", "android")
	javaHome := filepath.Join(path, "resources", "java")

	buildPath := filepath.Join(homePath, product+"_"+channelId)
	gamePath := filepath.Join(homePath, "game_demo.apk")
	expandPath := filepath.Join(homePath, "channel")

	utils.Copy(filepath.Join(homePath, "sds.keystore"), filepath.Join(buildPath, "sds.keystore"))
	utils.Copy(filepath.Join(homePath, "ic_launcher"), filepath.Join(buildPath, "ic_launcher"))
	utils.Copy(filepath.Join(homePath, "sds.keystore"), filepath.Join(buildPath, "sds.keystore"))
	preParams := models.PreParams{
		JavaHome:    javaHome,
		AndroidHome: androidHome,
		BuildPath:   buildPath,
		Channel:     channel,
		ChannelId:   channelId,
		GamePath:    gamePath,
		ExpandPath:  expandPath,
	}
	packaging.Excute(&preParams, &ProgressImpl{}, &LogImpl{})
}

type ProgressImpl struct {
}

func (ProgressImpl) Progress(channelId string, num int) {
	log.Println("当前进度", strconv.Itoa(num))
}

type LogImpl struct {
}

func (LogImpl) Printf(str string, data ...any) {
	log.Printf(str, data)
}

func (LogImpl) Println(data ...any) {
	log.Println(data...)
}
