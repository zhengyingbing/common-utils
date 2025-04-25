package packaging

import (
	"fmt"
	"path/filepath"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	models2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/utils"
	"strings"
	"sync"
	"time"
)

var (
	homePath, buildPath, gameApk, channel, channelId, androidJar, apktool, baksmali, smaliJar, aapt2, apksigner, dx, zipalign, java, javac, jarsigner string
)

func Execute(params *models2.PreParams, progress models2.ProgressCallback, logger models2.LogCallback) {
	logger.LogInfo("开始打包")
	t0 := time.Now().Unix()
	//1.初始化
	logger.LogInfo("开始初始化")
	initData(params)
	t1 := t0
	t2 := time.Now().Unix()
	logger.LogInfo("初始化打包环境完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 5)

	gameDirPath := filepath.Join(homePath, "gameDir")
	apks := []string{channel, "fastsdk", "channel", "core"}
	//2.反编译
	logger.LogInfo("开始反编译apk")
	decodeApk(gameDirPath, apks, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("反编译完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 10)
	//3.拷贝
	logger.LogInfo("开始拷贝apkDir")
	copyApkDirs(gameDirPath, apks, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("资源拷贝完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 20)

	gamePath := filepath.Join(buildPath, "gameDir")

	//4.将母包smali2之后的文件合并到主smali中
	logger.LogInfo("开始母包smali合并")
	utils.MergeSmaliFiles(gamePath)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("母包smali合并完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 25)

	//5.修复母包attrs
	logger.LogInfo("开始修复母包attrs")
	utils.GameRepairStyleable(gamePath, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("母包attrs修复完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 30)

	//6.渠道合并，母包优先级高
	logger.LogInfo("开始包合并处理")
	utils.MergeApkDir(buildPath, channel, gamePath, "", logger)
	logger.LogDebug("渠道包合并完成")
	progress.Progress(channelId, 35)

	//7.fastsdk合并，fastsdk优先级高
	//logger.LogDebug("开始fastsdk合并")
	utils.MergeApkDir(buildPath, "fastsdk", gamePath, "smali,assets,lib,manifest", logger)
	logger.LogDebug("fastsdk合并完成")
	progress.Progress(channelId, 40)

	//8.jni合并，lib和smali是jni优先级高
	//logger.LogDebug("开始jni合并")
	utils.MergeApkDir(buildPath, "core", gamePath, "smali,lib", logger)

	logger.LogDebug("jni合并完成")
	//9.插件包合并
	logger.LogDebug("插件包合并完成")
	progress.Progress(channelId, 45)

	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("包合并完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 50)

	//10.资源替换
	logger.LogInfo("开始资源替换")
	//删除未兼容全部架构的动态库
	utils.DeleteInvalidLibs(gamePath)
	logger.LogDebug("so库处理完成")
	utils.ReplaceRes(params, buildPath, gamePath, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("资源替换，耗时", t2-t1, "秒")
	progress.Progress(channelId, 60)

	//11.res构建
	logger.LogInfo("开始res.zip构建")
	utils.BuildRes(aapt2, gamePath, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("res.zip构建，耗时", t2-t1, "秒")
	progress.Progress(channelId, 70)

	//12.R文件构建
	logger.LogInfo("开始R文件构建")
	utils.BuildRHoolai(aapt2, androidJar, javac, dx, java, baksmali, gamePath, channelId, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("R文件构建完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 75)

	//13.分包
	logger.LogInfo("开始smali分包")
	smaliMap := utils.SmaliMap(gamePath, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("smali分包完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 80)

	//14.R文件处理
	logger.LogInfo("开始R文件处理")
	replaceR(smaliMap, channelId, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("R文件处理完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 85)

	//15.dex构建
	logger.LogInfo("开始dex构建")
	utils.BuildDex(java, smaliJar, gamePath, channelId, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("dex构建完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 90)

	//16.apk构建
	logger.LogInfo("开始apk构建")
	utils.CreateApk(gamePath, homePath, java, apktool, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("apk构建完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 95)

	//17.签名对齐
	logger.LogInfo("开始apk签名对齐处理")
	utils.SignApk(gamePath, jarsigner, apksigner, zipalign, params, logger)

	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("apk签名对齐完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 100)

	logger.LogInfo("打包成功！总耗时", t2-t0, "秒")

}

func replaceR(smaliMap map[string]string, channelId string, logger models2.LogCallback) {
	packageName := models2.GetServerDynamic(channelId)[models2.BundleId]
	packagePath := strings.Replace(packageName, ".", utils2.Symbol(), -1)
	for path, name := range smaliMap {
		if strings.HasPrefix(name, "R$") {
			if strings.Contains(path, packagePath) {
				continue
			} else {
				utils.ReplaceRFile(packagePath, path, name, logger)
			}
		}
	}
}

// 并发执行解包
func decodeApk(gameDirPath string, apks []string, logger models2.LogCallback) {

	target := filepath.Join(homePath, "target")
	shell := java + " -jar " + apktool + " --frame-path " + target + " --advance d %v" + " --only-main-classes -f -o %v"

	var wg sync.WaitGroup
	wg.Add(1)
	go decompile(&wg, shell, gameApk, gameDirPath, logger)

	shell = java + " -jar " + apktool + " --advance d %v" + " --only-main-classes -f -o %v"
	for _, apk := range apks {
		apkPath := filepath.Join(homePath, "channel", apk+".apk")
		dirPath := filepath.Join(homePath, "channel", apk+"Dir")
		wg.Add(1)
		go decompile(&wg, shell, apkPath, dirPath, logger)

	}
	wg.Wait()
}

// 并发执行拷贝
func copyApkDirs(gameDirPath string, apks []string, logger models2.LogCallback) {
	copyWg := sync.WaitGroup{}
	copyWg.Add(len(apks) + 1) // 所有解包任务 + 母包拷贝
	// 拷贝母包
	go func() {
		defer copyWg.Done()
		utils2.Copy(gameDirPath, filepath.Join(buildPath, "gameDir"), false)
		logger.LogDebug("gameDir拷贝完成")
	}()

	for _, apk := range apks {
		go func(name string) {
			defer copyWg.Done()
			dirPath := filepath.Join(homePath, "channel", name+"Dir")
			dstPath := filepath.Join(buildPath, name+"Dir")
			utils2.Copy(dirPath, dstPath, false)
			logger.LogDebug(name+"Dir", "拷贝完成")
		}(apk)
	}
	copyWg.Wait()
}

func decompile(wg *sync.WaitGroup, shell string, apkPath string, outPath string, logger models2.LogCallback) {
	defer wg.Done()
	if !utils2.Exist(filepath.Join(outPath, "AndroidManifest.xml")) {
		decodeShell := fmt.Sprintf(shell, apkPath, outPath)
		logger.LogDebug("执行命令：" + decodeShell)
		_ = utils2.ExecuteShell(decodeShell)
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
