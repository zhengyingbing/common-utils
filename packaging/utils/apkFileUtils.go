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
func MergeApkDir(buildPath, pluginName, gamePath string, rules string, logger models2.LogCallback, progress models2.ProgressCallback) {
	pluginPath := filepath.Join(buildPath, pluginName+"Dir")
	entries, err := os.ReadDir(pluginPath)
	if err != nil {
		logger.Println("Error: " + err.Error())
		panic(err.(interface{}))
	}
	if utils2.Exist(filepath.Join(pluginPath, "META-INF")) {
		utils2.Remove(filepath.Join(pluginPath, "META-INF"))
	}

	//logger.Printf("----开始%s合并", "smali")
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "smali") {
			priority := strings.Contains(rules, entry.Name())
			//smali合并时全部合并到母包的“主smali”中
			err = utils2.Move(filepath.Join(pluginPath, entry.Name()), filepath.Join(gamePath, "smali"), priority)
		}
	}
	//logger.Printf("----%s合并完成", "smali")

	logger.Printf("----开始%s合并", "res")
	if utils2.Exist(filepath.Join(pluginPath, "res")) {
		priority := strings.Contains(rules, "res")
		//res合并时需要进行特殊处理
		err = MergeRes(filepath.Join(pluginPath), pluginName, filepath.Join(gamePath), priority)
	}
	logger.Printf("----%s合并完成", "res")

	//logger.Printf("----开始%s合并", "AndroidManifest")
	if utils2.Exist(filepath.Join(pluginPath, "AndroidManifest.xml")) {
		err = MergeManifest(filepath.Join(pluginPath, "AndroidManifest.xml"), filepath.Join(gamePath, "AndroidManifest.xml"), logger)
	}
	//logger.Printf("----%s合并完成", "AndroidManifest")

	//logger.Printf("----开始%s合并", "yaml")
	if utils2.Exist(filepath.Join(pluginPath, "apktool.yml")) {
		err = MergeYaml(filepath.Join(pluginPath, "apktool.yml"), filepath.Join(gamePath, "apktool.yml"))
	}
	//logger.Printf("----%s合并完成", "yaml")

	//logger.Printf("----开始%s合并", "assets, lib, unknown, kotlin, original等")
	//最后合并assets, lib, unknown, kotlin, original等
	for _, entry := range entries {
		if entry.IsDir() && !strings.Contains(entry.Name(), "smali") && !strings.Contains(entry.Name(), "res") {
			priority := strings.Contains(rules, entry.Name())
			err = utils2.Move(filepath.Join(pluginPath, entry.Name()), filepath.Join(gamePath, entry.Name()), priority)
		}
	}
	//logger.Printf("----%s合并完成", "assets, lib, unknown, kotlin, original等")

	if err != nil {
		logger.Println("Error: " + err.Error())
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

func GameRepairStyleable(gameDir string) {
	println("开始执行GameRepairStyleable")
	attrsPath := filepath.Join(gameDir, "res", "values", "attrs.xml")
	newAttrsPath := filepath.Join(gameDir, "res", "values", "values_attrs.xml")

	if utils2.Exist(attrsPath) && !utils2.Exist(newAttrsPath) {
		pkgName := PackageName(filepath.Join(gameDir, "AndroidManifest.xml"))
		pkgPath := strings.Replace(pkgName, ".", utils2.Symbol(), -1)
		styleablePath := filepath.Join(gameDir, "smali", utils2.Symbol(), pkgPath, "R$styleable.smali")
		publicPath := filepath.Join(gameDir, "res", "values", "public.xml")
		RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrsPath)
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

	logger.Println("start build outapk!")
	packageName := models2.GetServerDynamic(channelId)[models2.BundleId]
	minSdkVersion := "21"
	targetSdkVersion := models2.GetServerDynamic(channelId)[models2.TargetSdkVersion]

	shellOutPutApk := strings.Join([]string{aapt2, "link", "-o", filepath.Join(gamePath, "output.apk"), "-I", androidJar, filepath.Join(gamePath, "res.zip"),
		"--manifest", filepath.Join(gamePath, "AndroidManifest.xml"), "--java", filepath.Join(gamePath, "gen"), "--custom-package",
		packageName, "--min-sdk-version", minSdkVersion, "--target-sdk-version", targetSdkVersion}, " ")

	err := utils2.ExecuteShell(shellOutPutApk)
	if err != nil {
		logger.Println("build R java failed, err: " + err.Error())
		panic(errors.New("build R java failed, err: " + err.Error()))
	} else {
		logger.Println("build R java success!")
	}

	genBinPath := filepath.Join(gamePath, "gen", "bin")
	err = utils2.CreateDir(genBinPath)
	if err != nil {
		println(genBinPath, "创建gen/bin目录失败：", err.Error())
	}

	rJavaPath := filepath.Join(gamePath, "gen", strings.ReplaceAll(packageName, ".", utils2.Symbol()), "R.java")
	shellRClass := strings.Join([]string{javac, "-encoding UTF-8 -target 1.8 -source 1.8 -bootclasspath", androidJar,
		"-d", genBinPath, rJavaPath}, " ")
	err = utils2.ExecuteShell(shellRClass)
	if err != nil {
		logger.Println("build R.class failed, err: " + err.Error())
		panic(errors.New("build R.class failed, err: " + err.Error()))
	} else {
		logger.Println("build R.class success!")
	}

	rHoolaiJavaPath := filepath.Join(gamePath, "gen", strings.ReplaceAll(packageName, ".", utils2.Symbol()), "R_Hoolai.java")
	_ = utils2.Copy(rJavaPath, rHoolaiJavaPath, true)
	utils2.ReplaceFile(rHoolaiJavaPath, "final class R", "class R_Hoolai")
	utils2.ReplaceFile(rHoolaiJavaPath, "final class", "class")

	shellRHoolaiClass := strings.Join([]string{javac, "-encoding UTF-8 -target 1.8 -source 1.8 -bootclasspath", androidJar,
		"-d", genBinPath, rHoolaiJavaPath}, " ")
	err = utils2.ExecuteShell(shellRHoolaiClass)
	if err != nil {
		logger.Println("build R_Hoolai.class failed, err: " + err.Error())
		panic(errors.New("build R_Hoolai.class failed, err: " + err.Error()))
	} else {
		logger.Println("build R_Hoolai.class success!")
	}

	println("start build dex")
	shellBuildDex := strings.Join([]string{dx, "--dex --output=", filepath.Join(genBinPath, "classes.dex"), genBinPath}, " ")
	err = utils2.ExecuteShell(shellBuildDex)
	if err != nil {
		logger.Println("build R_Hoolai.dex failed, err: " + err.Error())
		panic(errors.New("build R_Hoolai.dex failed, err: " + err.Error()))
	} else {
		logger.Println("build R_Hoolai.dex success!")
	}

	println("start decode dex")
	shellDecodeDex := strings.Join([]string{java, "-jar", baksmali, "d", filepath.Join(genBinPath, "classes.dex"), "-o",
		filepath.Join(genBinPath, "smali")}, " ")
	err = utils2.ExecuteShell(shellDecodeDex)
	if err != nil {
		logger.Println("decode R_Hoolai.dex failed, err: " + err.Error())
		panic(errors.New("decode R_Hoolai.dex failed, err: " + err.Error()))
	} else {
		logger.Println("decode R_Hoolai.dex success!")
	}

	err = utils2.Move(filepath.Join(genBinPath, "smali"), filepath.Join(gamePath, "smali"), true)
	if err != nil {
		logger.Println("move R_Hoolai.smali failed, err: " + err.Error())
		panic(errors.New("move R_Hoolai.smali failed, err: " + err.Error()))
	} else {
		logger.Println("move R_Hoolai.smali success!")
	}
}
