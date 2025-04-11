package packaging

import (
	"fmt"
	"path/filepath"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	models2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/utils"
	"sync"
	"time"
)

var (
	homePath, buildPath, gameApk, channel, channelId, androidJar, apktool, baksmali, smaliJar, aapt2, apksigner, dx, zipalign, java, javac, jarsigner string
)

func Execute(params *models2.PreParams, progress models2.ProgressCallback, logger models2.LogCallback) {
	initData(params)
	t0 := time.Now().Unix()
	progress.Progress(channelId, 1)
	//反编译+资源拷贝
	decodeAsyncAndCopy(logger, progress)
	gamePath := filepath.Join(buildPath, "gameDir")
	//将母包smali2之后的文件合并到主smali中
	utils.MergeSmaliFiles(filepath.Join(buildPath, "gameDir"))
	//修复attrs
	utils.GameRepairStyleable(gamePath)
	progress.Progress(channelId, 20)
	t1 := time.Now().Unix()
	logger.Println("合并前准备工作完成，耗时", t1-t0, "秒")

	//渠道合并，母包优先级高
	logger.Println("开始渠道包合并")
	utils.MergeApkDir(buildPath, channel, gamePath, "", logger, progress)
	progress.Progress(channelId, 35)
	logger.Println("渠道包合并完成")

	//fastsdk合并，fastsdk优先级高
	//logger.Println("开始fastsdk合并")
	utils.MergeApkDir(buildPath, "fastsdk", gamePath, "smali,assets,lib,manifest", logger, progress)
	progress.Progress(channelId, 40)
	//logger.Println("fastsdk合并完成")

	//jni合并，lib和smali是jni优先级高
	//logger.Println("开始jni合并")
	utils.MergeApkDir(buildPath, "core", gamePath, "smali,lib", logger, progress)
	progress.Progress(channelId, 45)
	//logger.Println("jni合并完成")

	t2 := time.Now().Unix()
	logger.Println("包合并完成，耗时", t2-t1, "秒")

	//删除未兼容全部架构的动态库
	utils.DeleteInvalidLibs(gamePath)
	utils.ReplaceRes(params, buildPath, gamePath, logger)
	progress.Progress(channelId, 60)
	t3 := time.Now().Unix()
	logger.Println("资源替换，耗时", t3-t2, "秒")

	utils.BuildRes(aapt2, gamePath, logger)
	progress.Progress(channelId, 70)
	t4 := time.Now().Unix()
	logger.Println("res.zip构建，耗时", t4-t3, "秒")
	utils.BuildRHoolai(aapt2, androidJar, javac, dx, java, baksmali, gamePath, channelId, logger)
}

// 串行执行解包
func decode(logger models2.LogCallback) {
	tm := time.Now().Unix()
	logger.Println("开始解包")
	gameDirPath := filepath.Join(homePath, "gameDir")
	target := filepath.Join(homePath, "target")
	shell := java + " -jar " + apktool + " --frame-path " + target + " --advance d %v" + " --only-main-classes -f -o %v"
	if !utils2.Exist(filepath.Join(gameDirPath, "AndroidManifest.xml")) {
		_ = utils2.ExecuteShell(fmt.Sprintf(shell, gameApk, gameDirPath))
	}
	shell = java + " -jar " + apktool + " --advance d %v" + " --only-main-classes -f -o %v"

	channelApk := filepath.Join(homePath, "channel", channel+".apk")
	channelDirPath := filepath.Join(homePath, "channel", channel+"Dir")
	if !utils2.Exist(filepath.Join(channelDirPath, "AndroidManifest.xml")) {
		_ = utils2.ExecuteShell(fmt.Sprintf(shell, channelApk, channelDirPath))
	}

	fastSdkApk := filepath.Join(homePath, "channel", "fastsdk.apk")
	fastSdkDirPath := filepath.Join(homePath, "channel", "fastsdkDir")
	if !utils2.Exist(filepath.Join(fastSdkDirPath, "AndroidManifest.xml")) {
		_ = utils2.ExecuteShell(fmt.Sprintf(shell, fastSdkApk, fastSdkDirPath))
	}

	jniApk := filepath.Join(homePath, "channel", "core.apk")
	jniDirPath := filepath.Join(homePath, "channel", "coreDir")
	if !utils2.Exist(filepath.Join(jniDirPath, "AndroidManifest.xml")) {
		_ = utils2.ExecuteShell(fmt.Sprintf(shell, jniApk, jniDirPath))
	}

	tm2 := time.Now().Unix() - tm
	logger.Println("解包完成，耗时：", tm2, "秒")
	_ = utils2.Copy(filepath.Join(homePath, "gameDir"), filepath.Join(buildPath, "gameDir"), true)
}

// 并发执行解包
func decodeAsyncAndCopy(logger models2.LogCallback, progress models2.ProgressCallback) {
	tm := time.Now().Unix()
	logger.Println("开始解包")
	gameDirPath := filepath.Join(homePath, "gameDir")
	target := filepath.Join(homePath, "target")
	shell := java + " -jar " + apktool + " --frame-path " + target + " --advance d %v" + " --only-main-classes -f -o %v"
	if !utils2.Exist(filepath.Join(gameDirPath, "AndroidManifest.xml")) {
		_ = utils2.ExecuteShell(fmt.Sprintf(shell, gameApk, gameDirPath))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go decompile(&wg, shell, gameApk, gameDirPath)

	shell = java + " -jar " + apktool + " --advance d %v" + " --only-main-classes -f -o %v"

	channelApk := filepath.Join(homePath, "channel", channel+".apk")
	channelDirPath := filepath.Join(homePath, "channel", channel+"Dir")
	wg.Add(1)
	go decompile(&wg, shell, channelApk, channelDirPath)

	fastSdkApk := filepath.Join(homePath, "channel", "fastsdk.apk")
	fastSdkDirPath := filepath.Join(homePath, "channel", "fastsdkDir")
	wg.Add(1)
	go decompile(&wg, shell, fastSdkApk, fastSdkDirPath)

	jniApk := filepath.Join(homePath, "channel", "core.apk")
	jniDirPath := filepath.Join(homePath, "channel", "coreDir")
	wg.Add(1)
	go decompile(&wg, shell, jniApk, jniDirPath)

	wg.Wait()
	progress.Progress(channelId, 20)
	tm2 := time.Now().Unix()
	logger.Println("解包完成，耗时：", tm2-tm, "秒")
	logger.Println("开始拷贝资源")

	copyApkDir(gameDirPath, filepath.Join(buildPath, "gameDir"))
	copyApkDir(channelDirPath, filepath.Join(buildPath, channel+"Dir"))
	copyApkDir(fastSdkDirPath, filepath.Join(buildPath, "fastsdkDir"))
	copyApkDir(jniDirPath, filepath.Join(buildPath, "coreDir"))
	tm3 := time.Now().Unix()
	logger.Println("资源拷贝完成，耗时：", tm3-tm2, "秒")
	progress.Progress(channelId, 30)
}

func copyApkDir(src, dst string) {
	utils2.Copy(src, dst, false)
}

func decompile(wg *sync.WaitGroup, shell string, apkPath string, outPath string) {
	defer wg.Done()
	if !utils2.Exist(filepath.Join(outPath, "AndroidManifest.xml")) {
		_ = utils2.ExecuteShell(fmt.Sprintf(shell, apkPath, outPath))
	}
}

func initData(params *models2.PreParams) {
	homePath = params.HomePath
	buildPath = params.BuildPath
	gameApk = params.GamePath
	channel = params.Channel
	channelId = params.ChannelId
	androidJar = filepath.Join(params.AndroidHome, "libs", "android.jar")
	apktool = filepath.Join(params.AndroidHome, "libs", "apktool.jar")
	baksmali = filepath.Join(params.AndroidHome, "libs", "baksmali.jar")
	smaliJar = filepath.Join(params.AndroidHome, "libs", "smali.jar")

	if utils2.CurrentOsType() == utils2.WINDOWS {
		aapt2 = filepath.Join(params.AndroidHome, "windows", "aapt2_64")
		apksigner = filepath.Join(params.AndroidHome, "windows", "apksigner.bat")
		dx = filepath.Join(params.AndroidHome, "windows", "dx.bat")
		zipalign = filepath.Join(params.AndroidHome, "windows", "zipalign.exe")

		java = filepath.Join(params.JavaHome, "win", "jre", "bin", "java")
		javac = filepath.Join(params.JavaHome, "win", "jre", "bin", "javac")
		jarsigner = filepath.Join(params.JavaHome, "win", "jre", "bin", "jarsigner.exe")
	} else if utils2.CurrentOsType() == utils2.MACOS {
		aapt2 = filepath.Join(params.AndroidHome, "macos", "aapt2_64")
		apksigner = filepath.Join(params.AndroidHome, "macos", "apksigner")
		dx = filepath.Join(params.AndroidHome, "macos", "dx")
		zipalign = filepath.Join(params.AndroidHome, "macos", "zipalign")

		java = "java"
		javac = "javac"
		jarsigner = "jarsigner"
	}

}
