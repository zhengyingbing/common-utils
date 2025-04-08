package main

import (
	"log"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/models"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging"
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
	cfg[models.AppName] = channel + "Demo"
	cfg[models.IconName] = "ic_launcher.png"
	cfg[models.TargetSdkVersion] = "30"
	cfg["dexMethodCounters"] = "60000"
	cfg[models.BundleId] = "com.hoolai.sf3.bytedance.gamecenter"
	cfg["appId"] = "614371"
	models.SetChannelDynamicConfig(channelId, cfg)
	androidHome := filepath.Join(path, "resources", "android")
	javaHome := filepath.Join(path, "resources", "java")

	buildPath := filepath.Join(homePath, product+"_"+channelId)
	gamePath := filepath.Join(homePath, "game_demo.apk")
	expandPath := filepath.Join(homePath, "channel")

	utils.Copy(filepath.Join(homePath, "access.config"), filepath.Join(buildPath, "access.config"))
	utils.Copy(filepath.Join(homePath, "ic_launcher.png"), filepath.Join(buildPath, "ic_launcher.png"))
	utils.Copy(filepath.Join(homePath, "aygd.keystore"), filepath.Join(buildPath, "aygd.keystore"))
	preParams := models.PreParams{
		JavaHome:    javaHome,
		AndroidHome: androidHome,
		BuildPath:   buildPath,
		Channel:     channel,
		ChannelId:   channelId,
		HomePath:    homePath,
		GamePath:    gamePath,
		ExpandPath:  expandPath,
	}
	packaging.Execute(&preParams, &ProgressImpl{}, &LogImpl{})
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
