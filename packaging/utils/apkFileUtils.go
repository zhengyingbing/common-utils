package utils

import (
	"errors"
	"github.com/zhengyingbing/common-utils/common/utils"
	"github.com/zhengyingbing/common-utils/packaging/models"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func CopyPlugins(srcDirs []string, buildPath []string, progress models.ProgressCallback) error {
	taskQueue := make(chan CopyTask, len(buildPath)*len(srcDirs))

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU()*2; i++ {
		wg.Add(1)
		go copyWorker(taskQueue, &wg)
	}
	for _, channel := range buildPath {
		for _, plugin := range srcDirs {
			taskQueue <- CopyTask{
				//ChannelId: channel,
				SrcPath: plugin,
				DstPath: channel,
			}
		}
	}
	close(taskQueue)
	wg.Wait()
	return nil
}

/**
 * rules：优先级，rules的值表示插件对应的模块优先级高
 * res目录合并之前确保smali已经合并完成，否则渠道中的styleable路径可能较难找到
 * 顺序：smali -> res -> AndroidManifest -> yaml -> other
 */
func MergeApkDir(buildPath, pluginName, gamePath string, rules string, logger models.LogCallback) {
	pluginPath := filepath.Join(buildPath, pluginName+"Dir")
	entries, err := os.ReadDir(pluginPath)
	if err != nil {
		logger.LogDebug("Error: " + err.Error())
		panic(err.(interface{}))
	}
	logger.LogDebug("pluginPath: " + pluginPath)
	MergeSmali(pluginPath, gamePath, pluginName, rules, entries, logger)

	if utils.Exist(filepath.Join(pluginPath, "res")) {
		priority := strings.Contains(rules, "res")
		//res合并时需要进行特殊处理
		err = MergeRes(filepath.Join(pluginPath), pluginName, filepath.Join(gamePath), priority, logger)
	}
	logger.LogVerbose(pluginName, "res合并完成")

	if utils.Exist(filepath.Join(pluginPath, "AndroidManifest.xml")) {
		err = MergeManifest(filepath.Join(pluginPath, "AndroidManifest.xml"), filepath.Join(gamePath, "AndroidManifest.xml"), logger)
	}
	logger.LogVerbose(pluginName, "AndroidManifest合并完成")

	if utils.Exist(filepath.Join(pluginPath, "apktool.yml")) {
		err = MergeYaml(filepath.Join(pluginPath, "apktool.yml"), filepath.Join(gamePath, "apktool.yml"))
	}
	logger.LogVerbose(pluginName, "yaml合并完成")

	//最后合并assets, lib, unknown, kotlin, original等
	for _, entry := range entries {
		if entry.IsDir() && !strings.Contains(entry.Name(), "smali") && !strings.Contains(entry.Name(), "res") {
			priority := strings.Contains(rules, entry.Name())
			err = utils.Move(filepath.Join(pluginPath, entry.Name()), filepath.Join(gamePath, entry.Name()), priority)
		}
	}
	logger.LogVerbose(pluginName, "assets, lib, unknown, kotlin, original等合并完成")

	if err != nil {
		logger.LogDebug("Error: " + err.Error())
		panic(err.(interface{}))
	}

}

func MergeSmaliFiles(src string) {
	entries, err := os.ReadDir(src)
	if err != nil {
		panic(err.(interface{}))
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "smali_") {
			utils.Move(filepath.Join(src, entry.Name()), filepath.Join(src, "smali"), false)
			utils.Remove(filepath.Join(src, entry.Name()))
		}
	}
}

/**
 * 动态库取交集
 */
func DeleteInvalidLibs(gamePath string) {
	entries, err := os.ReadDir(filepath.Join(gamePath, "lib"))
	if err != nil {
		panic(err.(interface{}))
	}
	max := 0
	for _, entry := range entries {
		if entry.IsDir() {
			entryFiles, err := os.ReadDir(filepath.Join(gamePath, "lib", entry.Name()))
			if err != nil {
				panic(err.(interface{}))
			}
			if max < len(entryFiles) {
				max = len(entryFiles)
			}
		}
	}
	for _, entry := range entries {
		if entry.IsDir() {
			entryFiles, err := os.ReadDir(filepath.Join(gamePath, "lib", entry.Name()))
			if err != nil {
				panic(err.(interface{}))
			}
			if max > len(entryFiles) {
				utils.Remove(filepath.Join(gamePath, "lib", entry.Name()))
			}
		}
	}
}

/**
 * 构建output.apk+R_hoolai.java -> 替换R_hoolai.java -> 生成R.class -> 生成R.dex -> 拷贝
 */
func BuildRHoolai(aapt2, androidJar, javac, dx, java, baksmali, gamePath, channelId string, logger models.LogCallback) {

	logger.LogDebug("start build outapk!")
	packageName := models.GetServerDynamic(channelId)[models.BundleId]
	minSdkVersion := "21"
	targetSdkVersion := models.GetServerDynamic(channelId)[models.TargetSdkVersion]

	shellOutPutApk := strings.Join([]string{aapt2, "link", "-o", filepath.Join(gamePath, "output.apk"), "-I", androidJar, filepath.Join(gamePath, "res.zip"),
		"--manifest", filepath.Join(gamePath, "AndroidManifest.xml"), "--java", filepath.Join(gamePath, "gen"), "--custom-package",
		packageName, "--min-sdk-version", minSdkVersion, "--target-sdk-version", targetSdkVersion}, " ")
	logger.LogDebug("执行命令：" + shellOutPutApk)
	err := utils.ExecuteShell(shellOutPutApk)
	if err != nil {
		logger.LogDebug("build R java failed, err: " + err.Error())
		panic(errors.New("build R java failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R java success!")
	}

	genBinPath := filepath.Join(gamePath, "gen", "bin")
	err = utils.CreateDir(genBinPath)
	if err != nil {
		logger.LogDebug(genBinPath, "创建gen/bin目录失败：", err.Error())
	}

	rJavaPath := filepath.Join(gamePath, "gen", strings.ReplaceAll(packageName, ".", utils.Symbol()), "R.java")
	shellRClass := strings.Join([]string{javac, "-encoding UTF-8 -target 1.8 -source 1.8 -bootclasspath", androidJar,
		"-d", genBinPath, rJavaPath}, " ")
	logger.LogDebug("执行命令：" + shellRClass)
	err = utils.ExecuteShell(shellRClass)
	if err != nil {
		logger.LogDebug("build R.class failed, err: " + err.Error())
		panic(errors.New("build R.class failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R.class success!")
	}

	rHoolaiJavaPath := filepath.Join(gamePath, "gen", strings.ReplaceAll(packageName, ".", utils.Symbol()), "R_hoolai.java")
	_ = utils.Copy(rJavaPath, rHoolaiJavaPath, true)
	utils.ReplaceFile(rHoolaiJavaPath, "final class R", "class R_hoolai")
	utils.ReplaceFile(rHoolaiJavaPath, "final class", "class")

	shellRHoolaiClass := strings.Join([]string{javac, "-encoding UTF-8 -target 1.8 -source 1.8 -bootclasspath", androidJar,
		"-d", genBinPath, rHoolaiJavaPath}, " ")
	logger.LogDebug("执行命令：" + shellRHoolaiClass)
	err = utils.ExecuteShell(shellRHoolaiClass)
	if err != nil {
		logger.LogDebug("build R_hoolai.class failed, err: " + err.Error())
		panic(errors.New("build R_hoolai.class failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R_hoolai.class success!")
	}

	logger.LogDebug("start build dex")
	shellBuildDex := strings.Join([]string{dx, "--dex --output=", filepath.Join(genBinPath, "classes.dex"), genBinPath}, " ")
	logger.LogDebug("执行命令：" + shellBuildDex)
	err = utils.ExecuteShell(shellBuildDex)
	if err != nil {
		logger.LogDebug("build R_hoolai.dex failed, err: " + err.Error())
		panic(errors.New("build R_hoolai.dex failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R_hoolai.dex success!")
	}

	logger.LogDebug("start decode dex")
	shellDecodeDex := strings.Join([]string{java, "-jar", baksmali, "d", filepath.Join(genBinPath, "classes.dex"), "-o",
		filepath.Join(genBinPath, "smali")}, " ")
	logger.LogDebug("执行命令：" + shellDecodeDex)
	err = utils.ExecuteShell(shellDecodeDex)
	if err != nil {
		logger.LogDebug("decode R_hoolai.dex failed, err: " + err.Error())
		panic(errors.New("decode R_hoolai.dex failed, err: " + err.Error()))
	} else {
		logger.LogDebug("decode R_hoolai.dex success!")
	}

	err = utils.Move(filepath.Join(genBinPath, "smali"), filepath.Join(gamePath, "smali"), true)
	if err != nil {
		logger.LogDebug("move R_hoolai.smali failed, err: " + err.Error())
		panic(errors.New("move R_hoolai.smali failed, err: " + err.Error()))
	} else {
		logger.LogDebug("move R_hoolai.smali success!")
	}
}

func CreateApk(gameDirPath, targetPath, java, apktool string, logger models.LogCallback) {
	logger.LogDebug("开始构建unsigned.apk")
	buildShell := strings.Join([]string{java, "-Dfile.encoding=utf-8 -jar", apktool, "--frame-path",
		targetPath, "b --use-aapt2", gameDirPath, "-o",
		filepath.Join(targetPath, "unsigned.apk")}, " ")
	logger.LogDebug("执行命令：" + buildShell)
	err := utils.ExecuteShell(buildShell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("构建unsigned.apk成功")
	}
}

func SignApk(gamePath, configPath, targetPath, outputApkPath, jarsigner, apksigner, zipalign string, params *models.PreParams, logger models.LogCallback) {
	logger.LogDebug("开始签名")

	signVersion := models.GetServerDynamic(params.ChannelId)[models.SignVersion]
	if strings.EqualFold(signVersion, "1") {
		v1Sign(gamePath, targetPath, outputApkPath, jarsigner, zipalign, params, logger)
	} else {
		v2Sign(gamePath, configPath, targetPath, outputApkPath, apksigner, zipalign, params, logger)
	}
}

func v2Sign(gamePath, configPath, targetPath, outputApkPath, apksigner, zipalign string, params *models.PreParams, logger models.LogCallback) {
	logger.LogDebug("apk开始v2+签名")
	keyStoreFile := filepath.Join(configPath, params.KeystoreName)
	alias := models.GetServerDynamic(params.ChannelId)[models.KeystoreAlias]
	keyStorePass := models.GetServerDynamic(params.ChannelId)[models.KeystorePass]
	pass := models.GetServerDynamic(params.ChannelId)[models.KeyPass]

	unSignedApk := filepath.Join(targetPath, "unsigned.apk")
	aligned := filepath.Join(targetPath, "aligned.apk")
	signedAlignApk := filepath.Join(targetPath, "signed_aligned.apk")

	zipApk(zipalign, unSignedApk, aligned, logger)

	v2Shell := strings.Join([]string{apksigner, "sign --v1-signing-enabled true --v2-signing-enabled true --v3-signing-enabled false --v4-signing-enabled false --ks",
		keyStoreFile, "--ks-key-alias", alias, ("--ks-pass pass:" + keyStorePass), ("--key-pass pass:" + pass), "-out",
		signedAlignApk, aligned}, " ")
	logger.LogDebug("执行命令：" + v2Shell)
	err := utils.ExecuteShell(v2Shell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("apk签名成功")
	}
	utils.Move(signedAlignApk, outputApkPath, true)
}

func v1Sign(gamePath, targetPath, outputApkPath, jarsigner, zipalign string, params *models.PreParams, logger models.LogCallback) {
	logger.LogDebug("apk开始v1签名")
	keyStoreFile := filepath.Join(gamePath, params.KeystoreName)
	alias := models.GetServerDynamic(params.ChannelId)[models.KeystoreAlias]
	keyStorePass := models.GetServerDynamic(params.ChannelId)[models.KeystorePass]
	pass := models.GetServerDynamic(params.ChannelId)[models.KeyPass]

	unSignedApk := filepath.Join(targetPath, "unsigned.apk")
	signedApk := filepath.Join(targetPath, "signed.apk")
	signedAlignApk := filepath.Join(targetPath, "signed_aligned.apk")

	v1Shell := strings.Join([]string{jarsigner, "-verbose -keystore", keyStoreFile, "-storepass", keyStorePass, "-keypass",
		pass, "-signedjar", signedApk, unSignedApk, alias, "-digestalg SHA1 -sigalg MD5withRSA"}, " ")
	logger.LogDebug("执行命令：" + v1Shell)
	err := utils.ExecuteShell(v1Shell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("apk签名成功")
	}

	zipApk(zipalign, signedApk, signedAlignApk, logger)
	utils.Move(signedAlignApk, outputApkPath, true)
}

func zipApk(zipalign, signedApk, signedAlignApk string, logger models.LogCallback) {
	logger.LogDebug("apk开始压缩")
	zipShell := strings.Join([]string{zipalign, "-f 4", signedApk, signedAlignApk}, " ")
	logger.LogDebug("执行命令：" + zipShell)
	err := utils.ExecuteShell(zipShell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("apk压缩成功")
	}
}
