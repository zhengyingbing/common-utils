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
	rootPath, apkPath, outputPath                                                                   string
	productId                                                                                       string
	androidJar, apktool, baksmali, smaliJar, aapt2, apksigner, dx, zipalign, java, javac, jarsigner string
)

func Preparation(productParam models.ProductParam, channelParams []models.ChannelParam, progress models.ProgressCallback) {

	//1.初始化
	utils.Write("开始初始化")
	gameDirPath := filepath.Join(rootPath, "gameDir")
	apks := []string{"fastsdk", "core"}
	//dstDirs := []string{}
	targetDirs := map[string]string{}

	dstDir := []string{}
	initData2(productParam)

	for _, param := range channelParams {
		channel := param.ChannelName
		channelId := param.ChannelId
		utils.Write("打包环境初始化成功！")
		utils.Write("产品参数：" + utils2.Struct2EscapeJson(productParam, false))
		buildPath := filepath.Join(productParam.RootPath, "build", productId+"_"+channelId) //C:\apktool\build\1-1
		utils2.CreateDir(buildPath)
		//dirsPath = append(dirsPath, buildPath)
		dstDir = append(dstDir, buildPath)
		apks = append(apks, channel)
		progress.Progress(param.ChannelId, 3)
		utils.Write("编译目录创建完成，进度完成4%")

		targetDirs[channelId] = buildPath
	}
	utils.Write("开始反编译, apks", apks)
	decodeApk(gameDirPath, apks)
	for _, param := range channelParams {
		utils.Write("apk包反编译完成，进度完成6%")
		progress.Progress(param.ChannelId, 5)
	}

	var srcPath []string
	srcPath = append(srcPath, filepath.Join(rootPath, "gameDir"))
	srcPath = append(srcPath, filepath.Join(rootPath, "sdk", "expand", "fastsdkDir"))
	srcPath = append(srcPath, filepath.Join(rootPath, "sdk", "expand", "coreDir"))
	utils.Write("开始拷贝源码")
	//utils.Write(fmt.Sprintf("%v", srcPath))
	//utils.Write(fmt.Sprintf("%v", dstDir))

	//utils.CopyPlugins(srcPath, dstDir, progress)
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
			utils.Write("母包源码拷贝完成，进度完成10%")
			progress.Progress(key, 15)
		}(dest)
	}
	go func() {
		wg.Wait()
	}()
	errChan := make(chan error, len(apks)*len(targetDirs))
	utils.Write("母包目录拷贝完成")

	sem := make(chan struct{}, 32)
	var wg2 sync.WaitGroup
	var completedNum = 1
	//var totalNum = len(apks)
	for _, apk := range apks {
		completedNum = completedNum + 1
		for _, dir := range targetDirs {
			wg2.Add(1)
			sem <- struct{}{}
			go func(src, destDir string, num int) {
				defer wg2.Done()
				defer func() { <-sem }()

				srcPath := filepath.Join("C:\\apktool", "sdk", "expand", src+"Dir")
				destPath := filepath.Join(destDir, src+"Dir")
				//if err := fastCopy(srcPath, destPath); err != nil {
				if err := utils2.Copy(srcPath, destPath, false); err != nil {
					errChan <- fmt.Errorf("%s -> %s 失败: %v", srcPath, destPath, err)
					return
				}
				//progress.Progress(key, completedNum/totalNum*15)
				//utils.Write(fmt.Sprintf("%s%s%d%s", apk, "源码拷贝完成，进度完成", num/totalNum*15, "%"))
				errChan <- nil
			}(apk, dir, completedNum)
		}
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
	utils.Write("拷贝完成，耗时", tm1-tm0, "秒")
	return nil
}

func Execute(param *models.PreParams, progress models.ProgressCallback, logger models.LogCallback) {

	//srcPath := filepath.Join(param.RootPath, "sdk", "expand", param.ChannelName+"Dir")
	//dstDir := param.BuildPath
	//utils.CopyPlugin(utils.CopyTask{SrcPath: srcPath, DstPath: dstDir})

	//srcPath := []string{filepath.Join(param.RootPath, "sdk", "expand", param.ChannelName+"Dir")}
	//dstDir := []string{param.BuildPath}
	//utils.CopyPlugins(srcPath, dstDir, progress)

	////utils2.Copy(srcPath, filepath.Join(param.BuildPath, param.ChannelName+"Dir"), true)
	//logger.LogInfo("渠道目录拷贝完成")
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
	logger.LogInfo("渠道参数准备完成！")
	logger.LogInfo("渠道参数：" + utils2.Struct2EscapeJson(param, false))
	utils2.Copy(filepath.Join(rootPath, "config", param.ChannelId, "access.config"), filepath.Join(param.BuildPath, "access.config"), true)
	logger.LogInfo("access.config拷贝完成")
	utils2.Copy(filepath.Join(rootPath, "config", param.ChannelId, "ic_launcher.png"), filepath.Join(param.BuildPath, "ic_launcher.png"), true)
	logger.LogInfo("应用icon拷贝完成")
	utils2.Copy(filepath.Join(rootPath, "config", param.ChannelId, "game.keystore"), filepath.Join(param.BuildPath, "game.keystore"), true)
	logger.LogInfo("签名文件拷贝完成")

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
	logger.LogInfo("母包smali合并完成，打包完成30%")
	progress.Progress(channelId, 30)

	//5.修复母包attrs
	logger.LogInfo("开始修复母包attrs")
	utils.RepairGameStyleable(gameDirPath, logger)

	logger.LogInfo("母包attrs修复完成")
	progress.Progress(channelId, 35)
	logger.LogInfo("母包attrs修复完成，打包完成35%")

	//6.渠道合并，母包优先级高
	logger.LogInfo("开始包合并处理")
	utils.MergeApkDir(buildPath, channel, gameDirPath, "", logger)
	logger.LogDebug("渠道包合并完成")
	progress.Progress(channelId, 40)
	logger.LogInfo("渠道包合并完成，打包完成40%")
	//7.fastsdk合并，fastsdk优先级高
	//logger.LogDebug("开始fastsdk合并")
	utils.MergeApkDir(buildPath, "fastsdk", gameDirPath, "smali,assets,lib,manifest", logger)
	logger.LogDebug("fastsdk合并完成")
	progress.Progress(channelId, 45)
	logger.LogInfo("fastsdk合并完成，打包完成45%")
	//8.jni合并，lib和smali是jni优先级高
	//logger.LogDebug("开始jni合并")
	utils.MergeApkDir(buildPath, "core", gameDirPath, "smali,lib", logger)
	logger.LogDebug("jni合并完成")
	logger.LogInfo("jni合并完成，打包完成50%")
	//9.插件包合并
	logger.LogDebug("插件包合并完成")

	logger.LogInfo("包合并完成")
	progress.Progress(channelId, 55)
	logger.LogInfo("插件包合并完成，打包完成55%")

	//10.资源替换
	logger.LogInfo("开始资源替换")
	//删除未兼容全部架构的动态库
	utils.DeleteInvalidLibs(gameDirPath)
	logger.LogDebug("so库处理完成")
	progress.Progress(channelId, 60)
	logger.LogInfo("so库处理完成，打包完成55%")

	configPath := filepath.Join(rootPath, "config", channel)
	utils.ReplaceRes(param, configPath, gameDirPath, logger)

	logger.LogInfo("资源替换")
	progress.Progress(channelId, 65)
	logger.LogInfo("资源替换完成，打包完成65%")

	//11.res构建
	logger.LogInfo("开始res.zip构建")
	utils.BuildRes(aapt2, gameDirPath, logger)

	logger.LogInfo("res.zip构建完成")
	progress.Progress(channelId, 70)
	logger.LogInfo("res.zip构建完成，打包完成70%")
	//12.R文件构建
	logger.LogInfo("开始R文件构建")
	utils.BuildRHoolai(aapt2, androidJar, javac, dx, java, baksmali, gameDirPath, channelId, logger)

	logger.LogInfo("R文件构建完成")
	progress.Progress(channelId, 75)
	logger.LogInfo("R文件构建完成，打包完成75%")

	//13.分包
	logger.LogInfo("开始smali分包")
	smaliMap := utils.SmaliMap(gameDirPath, logger)

	logger.LogInfo("smali分包完成")
	progress.Progress(channelId, 80)
	logger.LogInfo("smali分包完成，打包完成80%")
	//14.R文件处理
	logger.LogInfo("开始R文件处理")
	replaceR(smaliMap, channelId, logger)

	logger.LogInfo("R文件处理完成")
	progress.Progress(channelId, 85)
	logger.LogInfo("R文件处理完成，打包完成85%")

	//15.dex构建
	logger.LogInfo("开始dex构建")
	utils.BuildDex(java, smaliJar, gameDirPath, channelId, logger)

	logger.LogInfo("dex构建完成")
	progress.Progress(channelId, 90)
	logger.LogInfo("dex构建完成，打包完成90%")

	targetPath := filepath.Join(buildPath, "target")
	//16.apk构建
	logger.LogInfo("开始apk构建")
	utils.CreateApk(gameDirPath, targetPath, java, apktool, logger)

	logger.LogInfo("apk构建完成")
	progress.Progress(channelId, 95)
	logger.LogInfo("apk构建完成，打包完成95%")

	//17.签名对齐
	logger.LogInfo("开始apk签名对齐处理")

	outputApkPath := filepath.Join(outputPath, param.ApkName)
	logger.LogInfo("输出的文件名：", outputApkPath)
	utils.SignApk(gameDirPath, configPath, targetPath, outputApkPath, jarsigner, apksigner, zipalign, param, logger)

	logger.LogInfo("apk签名对齐完成")
	logger.LogInfo("apk签名对齐完成，打包完成100%")
	progress.Progress(channelId, 100)
	utils2.Remove(buildPath)
	logger.LogInfo("恭喜你打包成功！")
	logger.LogInfo("打包成功！")
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
func decodeApk(gameDirPath string, apks []string) []string {
	pluginDirsPath := []string{}
	utils.Write("母包目标目录：", gameDirPath)
	utils2.Remove(gameDirPath)
	target := filepath.Join(gameDirPath, "target")
	shell := java + " -jar " + apktool + " --frame-path " + target + " --advance d %v" + " --only-main-classes -f -o %v"

	var wg sync.WaitGroup
	wg.Add(1)
	go decompile(&wg, shell, apkPath, gameDirPath)

	shell = java + " -jar " + apktool + " --advance d %v" + " --only-main-classes -f -o %v"
	for _, apk := range apks {
		sdkPath := filepath.Join(rootPath, "sdk", apk+".apk")
		dirPath := filepath.Join(rootPath, "sdk", "expand", apk+"Dir")
		pluginDirsPath = append(pluginDirsPath, dirPath)
		if !utils2.Exist(dirPath) {
			utils2.CreateDir(dirPath)
		}
		wg.Add(1)
		go decompile(&wg, shell, sdkPath, dirPath)

	}
	wg.Wait()
	return pluginDirsPath
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

func decompile(wg *sync.WaitGroup, shell string, apkPath string, outPath string) {
	defer wg.Done()
	if !utils2.Exist(filepath.Join(outPath, "AndroidManifest.xml")) {
		decodeShell := fmt.Sprintf(shell, apkPath, outPath)
		utils.Write("执行命令：" + decodeShell)
		_ = utils2.ExecuteShell(decodeShell)
	}
}

func initData2(params models.ProductParam) {
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
