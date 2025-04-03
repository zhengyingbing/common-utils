package packaging

import (
	"os"
	"path/filepath"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/models"
	"strings"
)

func MergeApkDir(src, dst string, rules string, logger models.LogCallback, progress models.ProgressCallback) {
	entries, err := os.ReadDir(src)
	if err != nil {
		logger.Println("Error: " + err.Error())
		panic(err.(interface{}))
	}
	for _, entry := range entries {
		if entry.IsDir() {
			priority := strings.Contains(rules, entry.Name())

			if strings.Contains(entry.Name(), "smali") {
				//smali合并时全部合并到母包的“主smali”中
				err = utils2.MergeFile(filepath.Join(src, entry.Name()), filepath.Join(dst, "smali"), priority)
			} else if strings.Contains(entry.Name(), "res") {
				//res合并时需要进行特殊处理
				err = MergeRes(filepath.Join(src, entry.Name()), filepath.Join(dst, "res"), priority)
			} else {
				//assets,lib,unknown,kotlin,original等
				err = utils2.MergeFile(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()), priority)
			}
		} else {
			if strings.Contains(entry.Name(), "AndroidManifest.xml") {
				err = MergeManifest(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()), logger)
			} else if strings.Contains(entry.Name(), "apktool.yml") {
				err = MergeYaml(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()))
			}
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
	attrsPath := filepath.Join(gameDir, "res", "values", "attrs.xml")
	newAttrsPath := filepath.Join(gameDir, "res", "values", "values_attrs.xml")

	if utils2.Exist(attrsPath) && !utils2.Exist(newAttrsPath) {
		pkgName := PackageName(filepath.Join(gameDir, "AndroidManifest.xml"))
		pkgPath := strings.Replace(pkgName, ".", utils2.Space(), -1)
		styleablePath := filepath.Join(gameDir, "smali", pkgPath, "R$styleable.smali")
		publicPath := filepath.Join(gameDir, "res", "values", "public.xml")
		RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrsPath)
	}

}
