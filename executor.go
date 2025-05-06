// executor.go
//go:build !test

package main

import (
	"github.com/zhengyingbing/common-utils/common/utils"
	"github.com/zhengyingbing/common-utils/packaging"
	"github.com/zhengyingbing/common-utils/packaging/models"
	"log"
	"os"
	"path/filepath"

	"strconv"
)

//const (
//	product      = "3015"
//	channelId    = "10302"
//	channel      = "douyin"
//	game         = "aygd"
//	keystoreName = "aygd.keystore"
//keystoreName = "com.hoolai.sf3.bytedance.gamecenter"
//)

const (
	product      = "1"
	channelId    = "1"
	channel      = "hoolai"
	game         = "aygd"
	keystoreName = "aygd.keystore"
	packageName  = "com.hoolai.sdsxszycsds"
)

func main() {
	//path, _ := os.Getwd()
	path := "C:\\apktool"
	homePath := filepath.Join(path, "home")
	cfg := make(map[string]string)
	cfg[models.AppName] = channel + "Demo"
	cfg[models.IconName] = "ic_launcher.png"
	cfg[models.TargetSdkVersion] = "30"
	cfg[models.DexMethodCounters] = "60000"
	cfg[models.BundleId] = packageName
	cfg[models.Orientation] = "sensorPortrait"
	cfg[models.SignVersion] = "2"
	cfg[models.KeystoreAlias] = "aygd3"
	cfg[models.KeystorePass] = "aygd3123"
	cfg[models.KeyPass] = "aygd3123"
	cfg["appId"] = "614371"
	models.SetServerDynamic(channelId, cfg)
	androidHome := filepath.Join(path, "resources", "android")
	javaHome := filepath.Join(path, "resources", "java")

	buildPath := filepath.Join(homePath, product+"_"+channelId)
	gamePath := filepath.Join(homePath, "game_demo.apk")
	expandPath := filepath.Join(homePath, "channel")

	//remove(buildPath, filepath.Join(homePath, "temp"))
	utils.Remove(buildPath)
	utils.Copy(filepath.Join(homePath, channel, "access.config"), filepath.Join(buildPath, "access.config"), true)
	utils.Copy(filepath.Join(homePath, "ic_launcher.png"), filepath.Join(buildPath, "ic_launcher.png"), true)
	utils.Copy(filepath.Join(homePath, keystoreName), filepath.Join(buildPath, keystoreName), true)
	preParams := models.PreParams{
		JavaHome:     javaHome,
		AndroidHome:  androidHome,
		BuildPath:    buildPath,
		Channel:      channel,
		ChannelId:    channelId,
		HomePath:     homePath,
		GamePath:     gamePath,
		ExpandPath:   expandPath,
		KeystoreName: keystoreName,
	}
	packaging.Execute(&preParams, &ProgressImpl{}, &models.LogImpl{})
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
