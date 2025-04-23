package main

import (
	"path/filepath"
	models2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/utils"
)

const (
	macSplit     = "/"
	windowsSplit = "\\"
)

func main() {

	//err := utils.Move("C:\\apktool\\home\\3015_10302\\gameDir\\smali\\com\\hoolai\\demo\\CheckUtils.smali",
	//	"C:\\apktool\\home\\3015_10302\\gameDir\\smali_classes2\\com\\apmplus\\sdk\\cloudmessage\\1\\CheckUtils.smali", true)
	//if err != nil {
	//	println(err.Error())
	//}
	gamePath := "C:\\apktool\\home\\3015_10302"
	//java := "C:\\apktool\\resources\\java\\win\\jre\\bin\\java.exe"
	//apktool := "C:\\apktool\\resources\\android\\libs\\apktool.jar"
	//utils2.CreateApk(gamePath, java, apktool, &models2.LogImpl{})

	preParams := models2.PreParams{
		Channel:      "douyin",
		ChannelId:    "10302",
		GamePath:     gamePath,
		KeystoreName: "aygd.keystore",
	}
	cfg := make(map[string]string)
	cfg[models2.AppName] = "douyin" + "Demo"
	cfg[models2.IconName] = "ic_launcher.png"
	cfg[models2.TargetSdkVersion] = "30"
	cfg[models2.DexMethodCounters] = "60000"
	cfg[models2.BundleId] = "com.hoolai.sf3.bytedance.gamecenter"
	cfg[models2.SignVersion] = "2"
	cfg[models2.KeystoreAlias] = "aygd3"
	cfg[models2.KeystorePass] = "aygd3123"
	cfg[models2.KeyPass] = "aygd3123"
	cfg["appId"] = "614371"
	models2.SetServerDynamic("10302", cfg)
	jarsigner := filepath.Join("C:\\apktool\\resources\\java", "win", "jre", "bin", "jarsigner.exe")
	apksigner := filepath.Join("C:\\apktool\\resources\\android", "windows", "apksigner.bat")
	zipalign := filepath.Join("C:\\apktool\\resources\\android", "windows", "zipalign.exe")
	utils2.SignApk(gamePath, jarsigner, apksigner, zipalign, &preParams, &models2.LogImpl{})
}

type LoginCallback interface {
	OnSuccess(uid, token string)
	onFailed(err string)
}

type HandleLogin struct{}

func (h HandleLogin) OnSuccess(uid, token string) {

}
