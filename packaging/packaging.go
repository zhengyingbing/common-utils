package packaging

import (
	"fmt"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/models"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/utils"
)

var (
	channelId                                                                                       string
	androidJar, apktool, baksmali, smaliJar, aapt2, apksigner, dx, zipalign, java, javac, jarsigner string
)

func Excute(params *models.PreParams, progress models.ProgressCallback, logger models.LogCallback) {
	initData(params)
	progress.Progress(channelId, 1)
	decode(params, logger)
}

func decode(params *models.PreParams, logger models.LogCallback) {
	logger.Println("开始解包")
	gameApk := params.GamePath
	gameDirPath := filepath.Join(params.BuildPath, "gameDir")
	target := filepath.Join(params.BuildPath, "target")
	shell := java + " -jar " + apktool + " --frame-path " + target + " --advance d %v" + " --only-main-classes -f -o %v"
	if !utils.Exist(filepath.Join(gameDirPath, "AndroidManifest.xml")) {
		utils.ExecuteShell(fmt.Sprintf(shell, gameApk, gameDirPath))
	}
}

func initData(params *models.PreParams) {
	channelId = params.ChannelId
	androidJar = filepath.Join(params.AndroidHome, "libs", "android.jar")
	apktool = filepath.Join(params.AndroidHome, "libs", "apktool.jar")
	baksmali = filepath.Join(params.AndroidHome, "libs", "baksmali.jar")
	smaliJar = filepath.Join(params.AndroidHome, "libs", "smali.jar")

	if utils.CurrentOsType() == utils.WINDOWS {
		aapt2 = filepath.Join(params.AndroidHome, "windows", "aapt2_64")
		apksigner = filepath.Join(params.AndroidHome, "windows", "apksigner.bat")
		dx = filepath.Join(params.AndroidHome, "windows", "dx.bat")
		zipalign = filepath.Join(params.AndroidHome, "windows", "zipalign.exe")

		java = filepath.Join(params.JavaHome, "win", "jre", "bin", "java")
		javac = filepath.Join(params.JavaHome, "win", "jre", "bin", "javac")
		jarsigner = filepath.Join(params.JavaHome, "win", "jre", "bin", "jarsigner.exe")
	} else if utils.CurrentOsType() == utils.MACOS {
		aapt2 = filepath.Join(params.AndroidHome, "macos", "aapt2_64")
		apksigner = filepath.Join(params.AndroidHome, "macos", "apksigner")
		dx = filepath.Join(params.AndroidHome, "macos", "dx")
		zipalign = filepath.Join(params.AndroidHome, "macos", "zipalign")

		java = "java"
		javac = "javac"
		jarsigner = "jarsigner"
	}

}
