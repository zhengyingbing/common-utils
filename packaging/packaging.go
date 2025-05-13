package packaging

import (
	"fmt"
	utils2 "github.com/zhengyingbing/common-utils/common/utils"
	"github.com/zhengyingbing/common-utils/packaging/models"
	"github.com/zhengyingbing/common-utils/packaging/utils"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	rootPath, buildPath, apkPath, outputPath                                                        string
	productId, channel, channelId                                                                   string
	androidJar, apktool, baksmali, smaliJar, aapt2, apksigner, dx, zipalign, java, javac, jarsigner string
)

func Execute(params *models.PreParams, progress models.ProgressCallback, logger models.LogCallback) {
	logger.LogInfo("开始打包")
	t0 := time.Now().Unix()
	//1.初始化
	logger.LogInfo("开始初始化")
	initData(params)
	t1 := t0
	t2 := time.Now().Unix()
	logger.LogInfo("初始化打包环境完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 5)

	apks := []string{channel, "fastsdk", "core"}
	//2.反编译
	logger.LogInfo("开始反编译apk")
	decodeApk(filepath.Join(rootPath, "gameDir"), apks, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("反编译完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 10)
	//3.拷贝
	logger.LogInfo("开始拷贝apkDir", buildPath)
	gameDirPath := filepath.Join(buildPath, "gameDir")
	copyApkDirs(gameDirPath, apks, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("资源拷贝完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 20)

	//4.将母包smali2之后的文件合并到主smali中
	logger.LogInfo("开始母包smali合并")
	utils.MergeSmaliFiles(gameDirPath)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("母包smali合并完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 25)

	//5.修复母包attrs
	logger.LogInfo("开始修复母包attrs")
	utils.RepairGameStyleable(gameDirPath, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("母包attrs修复完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 30)

	//6.渠道合并，母包优先级高
	logger.LogInfo("开始包合并处理")
	utils.MergeApkDir(buildPath, channel, gameDirPath, "", logger)
	logger.LogDebug("渠道包合并完成")
	progress.Progress(channelId, 35)

	//7.fastsdk合并，fastsdk优先级高
	//logger.LogDebug("开始fastsdk合并")
	utils.MergeApkDir(buildPath, "fastsdk", gameDirPath, "smali,assets,lib,manifest", logger)
	logger.LogDebug("fastsdk合并完成")
	progress.Progress(channelId, 40)

	//8.jni合并，lib和smali是jni优先级高
	//logger.LogDebug("开始jni合并")
	utils.MergeApkDir(buildPath, "core", gameDirPath, "smali,lib", logger)

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
	utils.DeleteInvalidLibs(gameDirPath)
	logger.LogDebug("so库处理完成")

	configPath := filepath.Join(rootPath, "config", channelId)
	utils.ReplaceRes(params, configPath, gameDirPath, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("资源替换，耗时", t2-t1, "秒")
	progress.Progress(channelId, 60)

	//11.res构建
	logger.LogInfo("开始res.zip构建")
	utils.BuildRes(aapt2, gameDirPath, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("res.zip构建，耗时", t2-t1, "秒")
	progress.Progress(channelId, 70)

	//12.R文件构建
	logger.LogInfo("开始R文件构建")
	utils.BuildRHoolai(aapt2, androidJar, javac, dx, java, baksmali, gameDirPath, channelId, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("R文件构建完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 75)

	//13.分包
	logger.LogInfo("开始smali分包")
	smaliMap := utils.SmaliMap(gameDirPath, logger)
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
	utils.BuildDex(java, smaliJar, gameDirPath, channelId, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("dex构建完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 90)

	targetPath := filepath.Join(buildPath, "target")
	//16.apk构建
	logger.LogInfo("开始apk构建")
	utils.CreateApk(gameDirPath, targetPath, java, apktool, logger)
	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("apk构建完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 95)

	//17.签名对齐
	logger.LogInfo("开始apk签名对齐处理")

	outputApkPath := filepath.Join(outputPath, params.ApkName+".apk")
	utils.SignApk(gameDirPath, configPath, targetPath, outputApkPath, jarsigner, apksigner, zipalign, params, logger)

	t1 = t2
	t2 = time.Now().Unix()
	logger.LogInfo("apk签名对齐完成，耗时", t2-t1, "秒")
	progress.Progress(channelId, 100)

	logger.LogInfo("打包成功！总耗时", t2-t0, "秒")

}

func replaceR(smaliMap map[string]string, channelId string, logger models.LogCallback) {
	packageName := models.GetServerDynamic(channelId)[models.BundleId]
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
func decodeApk(gameDirPath string, apks []string, logger models.LogCallback) {

	target := filepath.Join(gameDirPath, "target")
	shell := java + " -jar " + apktool + " --frame-path " + target + " --advance d %v" + " --only-main-classes -f -o %v"

	var wg sync.WaitGroup
	wg.Add(1)
	go decompile(&wg, shell, apkPath, gameDirPath, logger)

	shell = java + " -jar " + apktool + " --advance d %v" + " --only-main-classes -f -o %v"
	for _, apk := range apks {
		apkPath := filepath.Join(rootPath, "sdk", apk+".apk")
		dirPath := filepath.Join(buildPath, apk+"Dir")
		wg.Add(1)
		go decompile(&wg, shell, apkPath, dirPath, logger)

	}
	wg.Wait()
}

// 并发执行拷贝
func copyApkDirs(gameDirPath string, apks []string, logger models.LogCallback) {
	copyWg := sync.WaitGroup{}
	copyWg.Add(len(apks) + 1) // 所有解包任务 + 母包拷贝
	// 拷贝母包
	go func() {
		defer copyWg.Done()
		utils2.Copy(filepath.Join(rootPath, "gameDir"), gameDirPath, false)
		logger.LogDebug("gameDir拷贝完成")
	}()

	for _, apk := range apks {
		go func(name string) {
			defer copyWg.Done()
			dirPath := filepath.Join(rootPath, "sdk", "expand", name+"Dir")
			dstPath := filepath.Join(buildPath, name+"Dir")
			utils2.Copy(dirPath, dstPath, false)
			logger.LogDebug(name+"Dir", "拷贝完成")
		}(apk)
	}
	copyWg.Wait()
}

func decompile(wg *sync.WaitGroup, shell string, apkPath string, outPath string, logger models.LogCallback) {
	defer wg.Done()
	if !utils2.Exist(filepath.Join(outPath, "AndroidManifest.xml")) {
		decodeShell := fmt.Sprintf(shell, apkPath, outPath)
		logger.LogDebug("执行命令：" + decodeShell)
		_ = utils2.ExecuteShell(decodeShell)
	}
}

func initData(params *models.PreParams) {
	fmt.Printf("%+v\n", params)
	productId = params.ProductId
	channel = params.ChannelName
	channelId = params.ChannelId

	rootPath = params.RootPath //C:\apktool
	if !utils2.Exist(rootPath) {
		utils2.CreateDir(rootPath)
	}
	buildPath = filepath.Join(params.RootPath, "build", productId+"_"+channelId) //C:\apktool\build\1-1
	if !utils2.Exist(buildPath) {
		utils2.CreateDir(buildPath)
	}
	outputPath = params.OutPutPath //C:\apktool\output
	if !utils2.Exist(outputPath) {
		utils2.CreateDir(outputPath)
	}
	apkPath = params.ApkPath

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
