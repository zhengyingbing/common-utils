package packaging

import (
	"os"
	"path/filepath"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/models"
	"strings"
)

/**
 * rules：优先级，rules的值表示插件对应的模块优先级高
 * res目录合并之前确保smali已经合并完成，否则渠道中的styleable路径可能较难找到
 * 顺序：smali -> res -> AndroidManifest -> yaml -> other
 */
func MergeApkDir(buildPath, pluginName, gamePath string, rules string, logger models.LogCallback, progress models.ProgressCallback) {
	pluginPath := filepath.Join(buildPath, pluginName+"Dir")
	entries, err := os.ReadDir(pluginPath)
	if err != nil {
		logger.Println("Error: " + err.Error())
		panic(err.(interface{}))
	}
	if utils2.Exist(filepath.Join(pluginPath, "META-INF")) {
		utils2.Remove(filepath.Join(pluginPath, "META-INF"))
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), "smali") {
			priority := strings.Contains(rules, entry.Name())
			//smali合并时全部合并到母包的“主smali”中
			err = utils2.MergeFile(filepath.Join(pluginPath, entry.Name()), filepath.Join(gamePath, "smali"), priority)
		}
	}

	if utils2.Exist(filepath.Join(pluginPath, "res")) {
		priority := strings.Contains(rules, "res")
		//res合并时需要进行特殊处理
		err = MergeRes(filepath.Join(pluginPath, "res"), pluginName, gamePath, priority)
	}

	if utils2.Exist(filepath.Join(pluginPath, "AndroidManifest.xml")) {
		err = MergeManifest(filepath.Join(pluginPath, "AndroidManifest.xml"), filepath.Join(gamePath, "AndroidManifest.xml"), logger)
	}

	if utils2.Exist(filepath.Join(pluginPath, "apktool.yml")) {
		err = MergeYaml(filepath.Join(pluginPath, "apktool.yml"), filepath.Join(gamePath, "apktool.yml"))
	}

	//最后合并assets, lib, unknown, kotlin, original等
	for _, entry := range entries {
		if entry.IsDir() && !strings.Contains(entry.Name(), "smali") && !strings.Contains(entry.Name(), "res") {
			priority := strings.Contains(rules, entry.Name())
			err = utils2.MergeFile(filepath.Join(pluginPath, entry.Name()), filepath.Join(gamePath, entry.Name()), priority)
		}
	}

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
			utils2.MergeFile(filepath.Join(src, entry.Name()), filepath.Join(src, "smali"), false)
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
		styleablePath := filepath.Join(gameDir, "smali", pkgPath, "R$styleable.smali")
		publicPath := filepath.Join(gameDir, "res", "values", "public.xml")
		RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrsPath)
	}

}

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
