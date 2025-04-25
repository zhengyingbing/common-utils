package utils

import (
	"errors"
	"os"
	"path/filepath"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	models2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strings"
)

/**
 * rules：优先级，rules的值表示插件对应的模块优先级高
 * res目录合并之前确保smali已经合并完成，否则渠道中的styleable路径可能较难找到
 * 顺序：smali -> res -> AndroidManifest -> yaml -> other
 */
func MergeApkDir(buildPath, pluginName, gamePath string, rules string, logger models2.LogCallback) {
	pluginPath := filepath.Join(buildPath, pluginName+"Dir")
	entries, err := os.ReadDir(pluginPath)
	if err != nil {
		logger.LogDebug("Error: " + err.Error())
		panic(err.(interface{}))
	}
	logger.LogDebug("pluginPath: " + pluginPath)
	MergeSmali(pluginPath, gamePath, pluginName, rules, entries, logger)

	if utils2.Exist(filepath.Join(pluginPath, "res")) {
		priority := strings.Contains(rules, "res")
		//res合并时需要进行特殊处理
		err = MergeRes(filepath.Join(pluginPath), pluginName, filepath.Join(gamePath), priority, logger)
	}
	logger.LogVerbose(pluginName, "res合并完成")

	if utils2.Exist(filepath.Join(pluginPath, "AndroidManifest.xml")) {
		err = MergeManifest(filepath.Join(pluginPath, "AndroidManifest.xml"), filepath.Join(gamePath, "AndroidManifest.xml"), logger)
	}
	logger.LogVerbose(pluginName, "AndroidManifest合并完成")

	if utils2.Exist(filepath.Join(pluginPath, "apktool.yml")) {
		err = MergeYaml(filepath.Join(pluginPath, "apktool.yml"), filepath.Join(gamePath, "apktool.yml"))
	}
	logger.LogVerbose(pluginName, "yaml合并完成")

	//最后合并assets, lib, unknown, kotlin, original等
	for _, entry := range entries {
		if entry.IsDir() && !strings.Contains(entry.Name(), "smali") && !strings.Contains(entry.Name(), "res") {
			priority := strings.Contains(rules, entry.Name())
			err = utils2.Move(filepath.Join(pluginPath, entry.Name()), filepath.Join(gamePath, entry.Name()), priority)
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
			utils2.Move(filepath.Join(src, entry.Name()), filepath.Join(src, "smali"), false)
			utils2.Remove(filepath.Join(src, entry.Name()))
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
				utils2.Remove(filepath.Join(gamePath, "lib", entry.Name()))
			}
		}
	}
}

/**
 * 构建output.apk+R_Hoolai.java -> 替换R_Hoolai.java -> 生成R.class -> 生成R.dex -> 拷贝
 */
func BuildRHoolai(aapt2, androidJar, javac, dx, java, baksmali, gamePath, channelId string, logger models2.LogCallback) {

	logger.LogDebug("start build outapk!")
	packageName := models2.GetServerDynamic(channelId)[models2.BundleId]
	minSdkVersion := "21"
	targetSdkVersion := models2.GetServerDynamic(channelId)[models2.TargetSdkVersion]

	shellOutPutApk := strings.Join([]string{aapt2, "link", "-o", filepath.Join(gamePath, "output.apk"), "-I", androidJar, filepath.Join(gamePath, "res.zip"),
		"--manifest", filepath.Join(gamePath, "AndroidManifest.xml"), "--java", filepath.Join(gamePath, "gen"), "--custom-package",
		packageName, "--min-sdk-version", minSdkVersion, "--target-sdk-version", targetSdkVersion}, " ")
	logger.LogDebug("执行命令：" + shellOutPutApk)
	err := utils2.ExecuteShell(shellOutPutApk)
	if err != nil {
		logger.LogDebug("build R java failed, err: " + err.Error())
		panic(errors.New("build R java failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R java success!")
	}

	genBinPath := filepath.Join(gamePath, "gen", "bin")
	err = utils2.CreateDir(genBinPath)
	if err != nil {
		logger.LogDebug(genBinPath, "创建gen/bin目录失败：", err.Error())
	}

	rJavaPath := filepath.Join(gamePath, "gen", strings.ReplaceAll(packageName, ".", utils2.Symbol()), "R.java")
	shellRClass := strings.Join([]string{javac, "-encoding UTF-8 -target 1.8 -source 1.8 -bootclasspath", androidJar,
		"-d", genBinPath, rJavaPath}, " ")
	logger.LogDebug("执行命令：" + shellRClass)
	err = utils2.ExecuteShell(shellRClass)
	if err != nil {
		logger.LogDebug("build R.class failed, err: " + err.Error())
		panic(errors.New("build R.class failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R.class success!")
	}

	rHoolaiJavaPath := filepath.Join(gamePath, "gen", strings.ReplaceAll(packageName, ".", utils2.Symbol()), "R_Hoolai.java")
	_ = utils2.Copy(rJavaPath, rHoolaiJavaPath, true)
	utils2.ReplaceFile(rHoolaiJavaPath, "final class R", "class R_Hoolai")
	utils2.ReplaceFile(rHoolaiJavaPath, "final class", "class")

	shellRHoolaiClass := strings.Join([]string{javac, "-encoding UTF-8 -target 1.8 -source 1.8 -bootclasspath", androidJar,
		"-d", genBinPath, rHoolaiJavaPath}, " ")
	logger.LogDebug("执行命令：" + shellRHoolaiClass)
	err = utils2.ExecuteShell(shellRHoolaiClass)
	if err != nil {
		logger.LogDebug("build R_Hoolai.class failed, err: " + err.Error())
		panic(errors.New("build R_Hoolai.class failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R_Hoolai.class success!")
	}

	logger.LogDebug("start build dex")
	shellBuildDex := strings.Join([]string{dx, "--dex --output=", filepath.Join(genBinPath, "classes.dex"), genBinPath}, " ")
	logger.LogDebug("执行命令：" + shellBuildDex)
	err = utils2.ExecuteShell(shellBuildDex)
	if err != nil {
		logger.LogDebug("build R_Hoolai.dex failed, err: " + err.Error())
		panic(errors.New("build R_Hoolai.dex failed, err: " + err.Error()))
	} else {
		logger.LogDebug("build R_Hoolai.dex success!")
	}

	logger.LogDebug("start decode dex")
	shellDecodeDex := strings.Join([]string{java, "-jar", baksmali, "d", filepath.Join(genBinPath, "classes.dex"), "-o",
		filepath.Join(genBinPath, "smali")}, " ")
	logger.LogDebug("执行命令：" + shellDecodeDex)
	err = utils2.ExecuteShell(shellDecodeDex)
	if err != nil {
		logger.LogDebug("decode R_Hoolai.dex failed, err: " + err.Error())
		panic(errors.New("decode R_Hoolai.dex failed, err: " + err.Error()))
	} else {
		logger.LogDebug("decode R_Hoolai.dex success!")
	}

	err = utils2.Move(filepath.Join(genBinPath, "smali"), filepath.Join(gamePath, "smali"), true)
	if err != nil {
		logger.LogDebug("move R_Hoolai.smali failed, err: " + err.Error())
		panic(errors.New("move R_Hoolai.smali failed, err: " + err.Error()))
	} else {
		logger.LogDebug("move R_Hoolai.smali success!")
	}
}

func CreateApk(gamePath, homePath, java, apktool string, logger models2.LogCallback) {
	logger.LogDebug("开始构建unsigned.apk")
	buildShell := strings.Join([]string{java, "-Dfile.encoding=utf-8 -jar", apktool, "--frame-path",
		filepath.Join(homePath, "target"), "b --use-aapt2", gamePath, "-o",
		filepath.Join(homePath, "target", "unsigned.apk")}, " ")
	logger.LogDebug("执行命令：" + buildShell)
	err := utils2.ExecuteShell(buildShell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("构建unsigned.apk成功")
	}
}

func SignApk(gamePath, jarsigner, apksigner, zipalign string, params *models2.PreParams, logger models2.LogCallback) {
	logger.LogDebug("开始签名")

	signVersion := models2.GetServerDynamic(params.ChannelId)[models2.SignVersion]
	if strings.EqualFold(signVersion, "1") {
		v1Sign(gamePath, jarsigner, zipalign, params, logger)
	} else {
		v2Sign(gamePath, apksigner, zipalign, params, logger)
	}
}

func v2Sign(gamePath, apksigner, zipalign string, params *models2.PreParams, logger models2.LogCallback) {
	logger.LogDebug("apk开始v2+签名")
	keyStoreFile := filepath.Join(params.HomePath, params.KeystoreName)
	alias := models2.GetServerDynamic(params.ChannelId)[models2.KeystoreAlias]
	keyStorePass := models2.GetServerDynamic(params.ChannelId)[models2.KeystorePass]
	pass := models2.GetServerDynamic(params.ChannelId)[models2.KeyPass]

	unSignedApk := filepath.Join(params.HomePath, "target", "unsigned.apk")
	aligned := filepath.Join(params.HomePath, "target", "aligned.apk")
	signedAlignApk := filepath.Join(params.HomePath, "target", "signed_aligned.apk")

	zipApk(zipalign, unSignedApk, aligned, logger)

	v2Shell := strings.Join([]string{apksigner, "sign --v1-signing-enabled true --v2-signing-enabled true --v3-signing-enabled false --v4-signing-enabled false --ks",
		keyStoreFile, "--ks-key-alias", alias, ("--ks-pass pass:" + keyStorePass), ("--key-pass pass:" + pass), "-out",
		signedAlignApk, aligned}, " ")
	logger.LogDebug("执行命令：" + v2Shell)
	err := utils2.ExecuteShell(v2Shell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("apk签名成功")
	}
}

func v1Sign(gamePath, jarsigner, zipalign string, params *models2.PreParams, logger models2.LogCallback) {
	logger.LogDebug("apk开始v1签名")
	keyStoreFile := filepath.Join(gamePath, params.KeystoreName)
	alias := models2.GetServerDynamic(params.ChannelId)[models2.KeystoreAlias]
	keyStorePass := models2.GetServerDynamic(params.ChannelId)[models2.KeystorePass]
	pass := models2.GetServerDynamic(params.ChannelId)[models2.KeyPass]

	unSignedApk := filepath.Join(params.HomePath, "target", "unsigned.apk")
	signedApk := filepath.Join(params.HomePath, "target", "signed.apk")
	signedAlignApk := filepath.Join(params.HomePath, "target", "signed_aligned.apk")

	v1Shell := strings.Join([]string{jarsigner, "-verbose -keystore", keyStoreFile, "-storepass", keyStorePass, "-keypass",
		pass, "-signedjar", signedApk, unSignedApk, alias, "-digestalg SHA1 -sigalg MD5withRSA"}, " ")
	logger.LogDebug("执行命令：" + v1Shell)
	err := utils2.ExecuteShell(v1Shell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("apk签名成功")
	}

	zipApk(zipalign, signedApk, signedAlignApk, logger)
}

func zipApk(zipalign, signedApk, signedAlignApk string, logger models2.LogCallback) {
	logger.LogDebug("apk开始压缩")
	zipShell := strings.Join([]string{zipalign, "-f 4", signedApk, signedAlignApk}, " ")
	logger.LogDebug("执行命令：" + zipShell)
	err := utils2.ExecuteShell(zipShell)
	if err != nil {
		panic(err.Error())
	} else {
		logger.LogDebug("apk压缩成功")
	}
}
