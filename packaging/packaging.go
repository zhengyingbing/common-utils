package packaging

import (
	_go "changeme/go"
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
	rootPath, apkPath, outputPath                                                                   string
	productId                                                                                       string
	androidJar, apktool, baksmali, smaliJar, aapt2, apksigner, dx, zipalign, java, javac, jarsigner string
)

var logManager *utils.LogManager

func Preparation(productParam _go.ProductParam, channelParams []_go.ChannelParam, progress models.ProgressCallback, logger models.LogCallback) {
	logManager = utils.NewLogManager(filepath.Join(productParam.RootPath, "logs"))
	defer logManager.Shutdown()
	logManager.StartFlusher(2 * time.Second)
	//1.初始化
	logger.LogInfo("开始初始化")
	gameDirPath := filepath.Join(rootPath, "gameDir")
	apks := []string{"fastsdk", "core"}
	targetDirs := map[string]string{}

	initData2(productParam)

	for _, param := range channelParams {
		channel := param.ChannelName
		channelId := param.ChannelId
		logManager.GetLogger(channelId).WriteLog("打包环境初始化成功！")
		logManager.GetLogger(channelId).WriteLog("产品参数：" + utils2.Struct2EscapeJson(productParam, false))
		buildPath := filepath.Join(productParam.RootPath, "build", productId+"_"+channelId) //C:\apktool\build\1-1
		utils2.Remove(buildPath)
		if !utils2.Exist(buildPath) {
			utils2.CreateDir(buildPath)
		}
		targetDirs[channelId] = buildPath
		apks = append(apks, channel)
		progress.Progress(param.ChannelId, 3)
		logManager.GetLogger(channelId).WriteLog("编译目录创建完成，进度完成4%")
	}
	logger.LogInfo("开始反编译, apks", apks)
	decodeApk(gameDirPath, apks, logger)
	for _, param := range channelParams {
		logManager.GetLogger(param.ChannelId).WriteLog("apk包反编译完成，进度完成6%")
		progress.Progress(param.ChannelId, 5)
	}

	logger.LogInfo("开始拷贝源码")
	CopyApk(gameDirPath, apks, targetDirs, progress)

}

// 拷贝母包、Fast、JNI
func CopyApk(gameDirPath string, apks []string, targetDirs map[string]string, progress models.ProgressCallback) error {
	var wg sync.WaitGroup
	tm0 := time.Now().Unix()
	for key, dest := range targetDirs {
		wg.Add(1)
		go func(dst string) {
			defer wg.Done()
			utils2.Copy(gameDirPath, filepath.Join(dst, "gameDir"), true)
			logManager.GetLogger(key).WriteLog("母包源码拷贝完成，进度完成10%")
			progress.Progress(key, 10)
		}(dest)
	}
	go func() {
		wg.Wait()
	}()
	errChan := make(chan error, len(apks)*len(targetDirs))
	println("母包目录拷贝完成")

	sem := make(chan struct{}, 32)
	var wg2 sync.WaitGroup
	var completedNum = 1
	var totalNum = len(apks)
	for _, apk := range apks {
		for key, dir := range targetDirs {
			wg2.Add(1)
			sem <- struct{}{}
			go func(src, destDir string) {
				defer wg2.Done()
				defer func() { <-sem }()

				srcPath := filepath.Join("C:\\apktool", "sdk", "expand", src+"Dir")
				destPath := filepath.Join(destDir, src+"Dir")
				//if err := fastCopy(srcPath, destPath); err != nil {
				if err := utils2.Copy(srcPath, destPath, false); err != nil {
					errChan <- fmt.Errorf("%s -> %s 失败: %v", srcPath, destPath, err)
					return
				}
				progress.Progress(key, completedNum/totalNum*15)
				logManager.GetLogger(key).WriteLog(fmt.Sprintf("%s%s%d%s", apk, "源码拷贝完成，进度完成", completedNum/totalNum*15, "%"))
				errChan <- nil
			}(apk, dir)
		}
		completedNum = completedNum + 1
	}
	go func() {
		wg2.Wait()
		close(errChan)
	}()
	// 检查错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	tm1 := time.Now().Unix()
	println("拷贝完成，耗时", tm1-tm0, "秒")
	return nil
}

func Execute(param *models.PreParams, progress models.ProgressCallback, logger models.LogCallback) {
	cfg := make(map[string]string)

	cfg[models.IconName] = "ic_launcher.png"
	cfg[models.TargetSdkVersion] = "30"
	cfg[models.DexMethodCounters] = "60000"
	cfg[models.AppName] = param.ProductName
	cfg[models.Orientation] = "sensorLandscape"
	cfg[models.SignVersion] = "2"
	cfg[models.KeystoreAlias] = "aygd3"
	cfg[models.KeystorePass] = "aygd3123"
	cfg[models.KeyPass] = "aygd3123"

	if strings.EqualFold(param.ChannelId, "10219") {
		cfg["appId"] = "614371"
	} else if strings.EqualFold(param.ChannelId, "10221") {
		cfg["appId"] = "2882303761520322194"
		cfg["appKey"] = "5642032276194"
	}

	cfg[models.BundleId] = param.PackageName
	models.SetServerDynamic(param.ChannelId, cfg)
	logManager.GetLogger(param.ChannelId).WriteLog("渠道参数准备完成！")
	logManager.GetLogger(param.ChannelId).WriteLog("渠道参数：" + utils2.Struct2EscapeJson(param, false))
	utils2.Copy(filepath.Join(rootPath, "config", param.ChannelId, "access.config"), filepath.Join(param.BuildPath, "access.config"), true)
	logManager.GetLogger(param.ChannelId).WriteLog("access.config拷贝完成")
	utils2.Copy(filepath.Join(rootPath, "config", param.ChannelId, "ic_launcher.png"), filepath.Join(param.BuildPath, "ic_launcher.png"), true)
	logManager.GetLogger(param.ChannelId).WriteLog("应用icon拷贝完成")
	utils2.Copy(filepath.Join(rootPath, "config", param.ChannelId, "game.keystore"), filepath.Join(param.BuildPath, "game.keystore"), true)
	logManager.GetLogger(param.ChannelId).WriteLog("签名文件拷贝完成")

	channelId := param.ChannelId
	channel := param.ChannelName
	buildPath := filepath.Join(param.RootPath, "build", productId+"_"+channelId)

	gameDirPath := filepath.Join(buildPath, "gameDir")

	logger.LogInfo("资源拷贝完成")
	//progress.Progress(channelId, 20)

	//4.将母包smali2之后的文件合并到主smali中
	logger.LogInfo("开始母包smali合并")
	utils.MergeSmaliFiles(gameDirPath)

	logger.LogInfo("母包smali合并完成")
	logManager.GetLogger(param.ChannelId).WriteLog("母包smali合并完成，打包完成30%")
	progress.Progress(channelId, 30)

	//5.修复母包attrs
	logger.LogInfo("开始修复母包attrs")
	utils.RepairGameStyleable(gameDirPath, logger)

	logger.LogInfo("母包attrs修复完成")
	progress.Progress(channelId, 35)
	logManager.GetLogger(param.ChannelId).WriteLog("母包attrs修复完成，打包完成35%")

	//6.渠道合并，母包优先级高
	logger.LogInfo("开始包合并处理")
	utils.MergeApkDir(buildPath, channel, gameDirPath, "", logger)
	logger.LogDebug("渠道包合并完成")
	progress.Progress(channelId, 40)
	logManager.GetLogger(param.ChannelId).WriteLog("渠道包合并完成，打包完成40%")
	//7.fastsdk合并，fastsdk优先级高
	//logger.LogDebug("开始fastsdk合并")
	utils.MergeApkDir(buildPath, "fastsdk", gameDirPath, "smali,assets,lib,manifest", logger)
	logger.LogDebug("fastsdk合并完成")
	progress.Progress(channelId, 45)
	logManager.GetLogger(param.ChannelId).WriteLog("fastsdk合并完成，打包完成45%")
	//8.jni合并，lib和smali是jni优先级高
	//logger.LogDebug("开始jni合并")
	utils.MergeApkDir(buildPath, "core", gameDirPath, "smali,lib", logger)
	logger.LogDebug("jni合并完成")
	logManager.GetLogger(param.ChannelId).WriteLog("jni合并完成，打包完成50%")
	//9.插件包合并
	logger.LogDebug("插件包合并完成")

	logger.LogInfo("包合并完成")
	progress.Progress(channelId, 55)
	logManager.GetLogger(param.ChannelId).WriteLog("插件包合并完成，打包完成55%")

	//10.资源替换
	logger.LogInfo("开始资源替换")
	//删除未兼容全部架构的动态库
	utils.DeleteInvalidLibs(gameDirPath)
	logger.LogDebug("so库处理完成")
	progress.Progress(channelId, 60)
	logManager.GetLogger(param.ChannelId).WriteLog("so库处理完成，打包完成55%")

	configPath := filepath.Join(rootPath, "config", channel)
	utils.ReplaceRes(param, configPath, gameDirPath, logger)

	logger.LogInfo("资源替换")
	progress.Progress(channelId, 65)
	logManager.GetLogger(param.ChannelId).WriteLog("资源替换完成，打包完成65%")

	//11.res构建
	logger.LogInfo("开始res.zip构建")
	utils.BuildRes(aapt2, gameDirPath, logger)

	logger.LogInfo("res.zip构建完成")
	progress.Progress(channelId, 70)
	logManager.GetLogger(param.ChannelId).WriteLog("res.zip构建完成，打包完成70%")
	//12.R文件构建
	logger.LogInfo("开始R文件构建")
	utils.BuildRHoolai(aapt2, androidJar, javac, dx, java, baksmali, gameDirPath, channelId, logger)

	logger.LogInfo("R文件构建完成")
	progress.Progress(channelId, 75)
	logManager.GetLogger(param.ChannelId).WriteLog("R文件构建完成，打包完成75%")

	//13.分包
	logger.LogInfo("开始smali分包")
	smaliMap := utils.SmaliMap(gameDirPath, logger)

	logger.LogInfo("smali分包完成")
	progress.Progress(channelId, 80)
	logManager.GetLogger(param.ChannelId).WriteLog("smali分包完成，打包完成80%")
	//14.R文件处理
	logger.LogInfo("开始R文件处理")
	replaceR(smaliMap, channelId, logger)

	logger.LogInfo("R文件处理完成")
	progress.Progress(channelId, 85)
	logManager.GetLogger(param.ChannelId).WriteLog("R文件处理完成，打包完成85%")

	//15.dex构建
	logger.LogInfo("开始dex构建")
	utils.BuildDex(java, smaliJar, gameDirPath, channelId, logger)

	logger.LogInfo("dex构建完成")
	progress.Progress(channelId, 90)
	logManager.GetLogger(param.ChannelId).WriteLog("dex构建完成，打包完成90%")

	targetPath := filepath.Join(buildPath, "target")
	//16.apk构建
	logger.LogInfo("开始apk构建")
	utils.CreateApk(gameDirPath, targetPath, java, apktool, logger)

	logger.LogInfo("apk构建完成")
	progress.Progress(channelId, 95)
	logManager.GetLogger(param.ChannelId).WriteLog("apk构建完成，打包完成95%")

	//17.签名对齐
	logger.LogInfo("开始apk签名对齐处理")

	outputApkPath := filepath.Join(outputPath, param.ApkName)
	println("输出的文件名：", outputApkPath)
	utils.SignApk(gameDirPath, configPath, targetPath, outputApkPath, jarsigner, apksigner, zipalign, param, logger)

	logger.LogInfo("apk签名对齐完成")
	logManager.GetLogger(param.ChannelId).WriteLog("apk签名对齐完成，打包完成100%")
	progress.Progress(channelId, 100)
	utils2.Remove(buildPath)
	logger.LogInfo("恭喜你打包成功！")
	logManager.GetLogger(param.ChannelId).WriteLog("打包成功！")
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
	println("母包目标目录：", gameDirPath)
	utils2.Remove(gameDirPath)
	target := filepath.Join(gameDirPath, "target")
	shell := java + " -jar " + apktool + " --frame-path " + target + " --advance d %v" + " --only-main-classes -f -o %v"

	var wg sync.WaitGroup
	wg.Add(1)
	go decompile(&wg, shell, apkPath, gameDirPath, logger)

	shell = java + " -jar " + apktool + " --advance d %v" + " --only-main-classes -f -o %v"
	for _, apk := range apks {
		sdkPath := filepath.Join(rootPath, "sdk", apk+".apk")
		dirPath := filepath.Join(rootPath, "sdk", "expand", apk+"Dir")
		if !utils2.Exist(dirPath) {
			utils2.CreateDir(dirPath)
		}
		wg.Add(1)
		go decompile(&wg, shell, sdkPath, dirPath, logger)

	}
	wg.Wait()
}

//// 并发执行拷贝
//func copyApkDirs(gameDirPath string, apks []string, logger models.LogCallback) {
//	copyWg := sync.WaitGroup{}
//	copyWg.Add(len(apks) + 1) // 所有解包任务 + 母包拷贝
//	// 拷贝母包
//	go func() {
//		defer copyWg.Done()
//		utils2.Copy(filepath.Join(rootPath, "gameDir"), gameDirPath, false)
//		logger.LogDebug("gameDir拷贝完成")
//	}()
//
//	for _, apk := range apks {
//		go func(name string) {
//			defer copyWg.Done()
//			dirPath := filepath.Join(rootPath, "sdk", "expand", name+"Dir")
//			dstPath := filepath.Join(buildPath, name+"Dir")
//			utils2.Copy(dirPath, dstPath, false)
//			logger.LogDebug(name+"Dir", "拷贝完成")
//		}(apk)
//	}
//	copyWg.Wait()
//}

func decompile(wg *sync.WaitGroup, shell string, apkPath string, outPath string, logger models.LogCallback) {
	defer wg.Done()
	if !utils2.Exist(filepath.Join(outPath, "AndroidManifest.xml")) {
		decodeShell := fmt.Sprintf(shell, apkPath, outPath)
		logger.LogDebug("执行命令：" + decodeShell)
		_ = utils2.ExecuteShell(decodeShell)
	}
}

func initData2(params _go.ProductParam) {
	productId = params.ProductId
	rootPath = params.RootPath //C:\apktool
	if !utils2.Exist(rootPath) {
		utils2.CreateDir(rootPath)
	}

	outputPath = params.OutputPath //C:\apktool\output
	if !utils2.Exist(outputPath) {
		utils2.CreateDir(outputPath)
	}
	apkPath = params.ApkPath

	androidJar = filepath.Join(params.AndroidPath, "libs", "android.jar")
	apktool = filepath.Join(params.AndroidPath, "libs", "apktool.jar")
	baksmali = filepath.Join(params.AndroidPath, "libs", "baksmali.jar")
	smaliJar = filepath.Join(params.AndroidPath, "libs", "smali.jar")

	if utils2.CurrentOsType() == utils2.WINDOWS {
		aapt2 = filepath.Join(params.AndroidPath, "windows", "aapt2_64")
		apksigner = filepath.Join(params.AndroidPath, "windows", "apksigner.bat")
		dx = filepath.Join(params.AndroidPath, "windows", "dx.bat")
		zipalign = filepath.Join(params.AndroidPath, "windows", "zipalign.exe")

		java = filepath.Join(params.JavaPath, "win", "jre", "bin", "java")
		javac = filepath.Join(params.JavaPath, "win", "jre", "bin", "javac")
		jarsigner = filepath.Join(params.JavaPath, "win", "jre", "bin", "jarsigner.exe")
	} else if utils2.CurrentOsType() == utils2.MACOS {
		aapt2 = filepath.Join(params.AndroidPath, "macos", "aapt2_64")
		apksigner = filepath.Join(params.AndroidPath, "macos", "apksigner")
		dx = filepath.Join(params.AndroidPath, "macos", "dx")
		zipalign = filepath.Join(params.AndroidPath, "macos", "zipalign")

		java = "java"
		javac = "javac"
		jarsigner = "jarsigner"
	}
}

//func initData(params *models.PreParams) {
//	fmt.Printf("%+v\n", params)
//	productId = params.ProductId
//	channel = params.ChannelName
//	channelId = params.ChannelId
//
//	rootPath = params.RootPath //C:\apktool
//	if !utils2.Exist(rootPath) {
//		utils2.CreateDir(rootPath)
//	}
//	buildPath = filepath.Join(params.RootPath, "build", productId+"_"+channelId) //C:\apktool\build\1-1
//	if !utils2.Exist(buildPath) {
//		utils2.CreateDir(buildPath)
//	}
//	outputPath = params.OutPutPath //C:\apktool\output
//	if !utils2.Exist(outputPath) {
//		utils2.CreateDir(outputPath)
//	}
//	apkPath = params.ApkPath
//
//	androidJar = filepath.Join(params.AndroidHome, "libs", "android.jar")
//	apktool = filepath.Join(params.AndroidHome, "libs", "apktool.jar")
//	baksmali = filepath.Join(params.AndroidHome, "libs", "baksmali.jar")
//	smaliJar = filepath.Join(params.AndroidHome, "libs", "smali.jar")
//
//	if utils2.CurrentOsType() == utils2.WINDOWS {
//		aapt2 = filepath.Join(params.AndroidHome, "windows", "aapt2_64")
//		apksigner = filepath.Join(params.AndroidHome, "windows", "apksigner.bat")
//		dx = filepath.Join(params.AndroidHome, "windows", "dx.bat")
//		zipalign = filepath.Join(params.AndroidHome, "windows", "zipalign.exe")
//
//		java = filepath.Join(params.JavaHome, "win", "jre", "bin", "java")
//		javac = filepath.Join(params.JavaHome, "win", "jre", "bin", "javac")
//		jarsigner = filepath.Join(params.JavaHome, "win", "jre", "bin", "jarsigner.exe")
//	} else if utils2.CurrentOsType() == utils2.MACOS {
//		aapt2 = filepath.Join(params.AndroidHome, "macos", "aapt2_64")
//		apksigner = filepath.Join(params.AndroidHome, "macos", "apksigner")
//		dx = filepath.Join(params.AndroidHome, "macos", "dx")
//		zipalign = filepath.Join(params.AndroidHome, "macos", "zipalign")
//
//		java = "java"
//		javac = "javac"
//		jarsigner = "jarsigner"
//	}
//
//}
