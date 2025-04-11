package main

import (
	"log"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging"
	models2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strconv"
)

const (
	product   = "3015"
	channelId = "10302"
	channel   = "douyin"
	game      = "aygd"
)

func main() {
	//path, _ := os.Getwd()
	path := "C:\\apktool"
	homePath := filepath.Join(path, "home")
	cfg := make(map[string]string)
	cfg[models2.AppName] = channel + "Demo"
	cfg[models2.IconName] = "ic_launcher.png"
	cfg[models2.TargetSdkVersion] = "30"
	cfg["dexMethodCounters"] = "60000"
	cfg[models2.BundleId] = "com.hoolai.sf3.bytedance.gamecenter"
	cfg["appId"] = "614371"
	models2.SetServerDynamic(channelId, cfg)
	androidHome := filepath.Join(path, "resources", "android")
	javaHome := filepath.Join(path, "resources", "java")

	buildPath := filepath.Join(homePath, product+"_"+channelId)
	gamePath := filepath.Join(homePath, "game_demo.apk")
	expandPath := filepath.Join(homePath, "channel")

	remove(buildPath, filepath.Join(homePath, "temp"))

	utils.Copy(filepath.Join(homePath, "access.config"), filepath.Join(buildPath, "access.config"), true)
	utils.Copy(filepath.Join(homePath, "ic_launcher.png"), filepath.Join(buildPath, "ic_launcher.png"), true)
	utils.Copy(filepath.Join(homePath, "aygd.keystore"), filepath.Join(buildPath, "aygd.keystore"), true)
	preParams := models2.PreParams{
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

func remove(src, dst string) error {
	os.Rename(src, dst)
	go func() {
		utils.Remove(dst)
	}()
	return nil
}

type ProgressImpl struct {
}

func (ProgressImpl) Progress(channelId string, num int) {
	log.Println("当前进度", strconv.Itoa(num)+"%")
}

type LogImpl struct {
}

func (LogImpl) Printf(str string, data ...any) {
	log.Printf(str, data)
}

func (LogImpl) Println(data ...any) {
	log.Println(data...)
}
