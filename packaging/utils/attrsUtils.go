package utils

import (
	xml "github.com/xyjwsj/xml_parser"
	"os"
	"path/filepath"
	"regexp"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	models2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strings"
)

func RepairGameStyleable(gameDir string, logger models2.LogCallback) {
	logger.LogDebug("开始执行GameRepairStyleable")
	publicPath := filepath.Join(gameDir, "res", "values", "public.xml")
	newAttrsPath := filepath.Join(gameDir, "res", "values", "values_attrs.xml")
	pkgName := PackageName(filepath.Join(gameDir, "AndroidManifest.xml"))
	pkgPath := strings.Replace(pkgName, ".", utils2.Symbol(), -1)
	//TODO: 原始包名和最终包名如果不一致，会找不到R$styleable.smali，因为此文件在原始包名路径下生成
	styleablePath := filepath.Join(gameDir, "smali", utils2.Symbol(), pkgPath, "R$styleable.smali")
	if utils2.Exist(styleablePath) {
		RebuildStyleable(styleablePath, newAttrsPath, logger)
	}
	err := utils2.Remove(publicPath)
	if err != nil {
		logger.LogDebug("删除", publicPath, err.Error())
	}
}

func RepairPluginStyleable(pluginPath, pluginName, gamePath string, logger models2.LogCallback) {
	publicPath := filepath.Join(pluginPath, "res", "values", "public.xml")
	newAttrsPath := filepath.Join(pluginPath, "res", "values", "values_attrs.xml")
	if utils2.Exist(newAttrsPath) || strings.EqualFold(pluginName, "firebase") {
		err := os.Remove(publicPath)
		if err != nil {
			return
		}
		return
	}
	packageName := PackageName(filepath.Join(pluginPath, "AndroidManifest.xml"))
	packagePath := strings.Replace(packageName, ".", utils2.Symbol(), -1)
	styleablePath := filepath.Join(gamePath, "smali", packagePath, "R$styleable.smali")

	if utils2.Exist(styleablePath) {
		logger.LogDebug(pluginName, "的styleablePath = ", styleablePath)
		RebuildStyleable(styleablePath, newAttrsPath, logger)
	}
	err := utils2.Remove(publicPath)
	if err != nil {
		logger.LogDebug("删除", publicPath, err.Error())
	}
}

func RebuildStyleable(styleablePath, newAttrPath string, logger models2.LogCallback) {
	//解析smali
	result := make(map[string][]string)
	pattern := regexp.MustCompile(`_([a-z]+)`)
	utils2.ReadLine(styleablePath, func(err error, line int, content string) bool {
		if content == "" {
			return false
		}
		content = strings.TrimSpace(content)
		if strings.Contains(content, ".field public static final ") && strings.Contains(content, ":I") {
			content = strings.Replace(content, ".field public static final ", "", -1)
			v := content[0:strings.Index(content, ":I")]

			println(v)
			matches := pattern.FindStringSubmatch(v)
			if len(matches) >= 0 {
				prefix := matches[1] // 前半部分（Aa_Bc）
				i := strings.Index(v, prefix)
				key := v[0 : i-1]
				value := v[i:]
				if strings.HasPrefix(value, "android_") {
					value = strings.Replace(value, "android_", "android:", -1)
				}
				//println("前半段：", key)
				//println("后半段:", value)
				result[key] = append(result[key], value)
			}
		}
		return false
	})
	//开始构建xml
	tag := xml.Tag{
		Name:      "resources",
		Attribute: nil,
		ChildTags: make([]*xml.Tag, 0),
	}
	for k, v := range result {
		parentTag := xml.Tag{
			Name:      "declare-styleable",
			Attribute: map[string]string{"name": k},
			ChildTags: make([]*xml.Tag, 0),
			Parent:    nil,
		}
		tag.ChildTags = append(tag.ChildTags, &parentTag)
		for _, v2 := range v {
			attribute := make(map[string]string)
			attribute["name"] = v2
			parentTag.ChildTags = append(parentTag.ChildTags, &xml.Tag{
				Name:      "attr",
				Attribute: attribute,
				ChildTags: nil,
				Parent:    nil,
			})
		}
	}
	xml.Serializer(tag, xml.XmlHeaderType, newAttrPath)
}

func BuildRes(aapt2, gamePath string, logger models2.LogCallback) {
	logger.LogDebug("开始构建res")
	shellString := aapt2 + " compile --dir " + filepath.Join(gamePath, "res") + " -o " + filepath.Join(gamePath, "res.zip")
	logger.LogDebug("执行命令:" + shellString)
	err := utils2.ExecuteShell(shellString)
	if err != nil {
		logger.LogDebug("res.zip build failed!")
		panic(err.Error())
	} else {
		logger.LogDebug("res.zip build success!")
	}
}
