package utils

import (
	xml "github.com/xyjwsj/xml_parser"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/smali"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	models2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strings"
)

func GameRepairStyleable(gameDir string, logger models2.LogCallback) {
	logger.LogDebug("开始执行GameRepairStyleable")
	attrsPath := filepath.Join(gameDir, "res", "values", "attrs.xml")
	newAttrsPath := filepath.Join(gameDir, "res", "values", "values_attrs.xml")

	if utils2.Exist(attrsPath) && !utils2.Exist(newAttrsPath) {
		pkgName := PackageName(filepath.Join(gameDir, "AndroidManifest.xml"))
		pkgPath := strings.Replace(pkgName, ".", utils2.Symbol(), -1)
		styleablePath := filepath.Join(gameDir, "smali", utils2.Symbol(), pkgPath, "R$styleable.smali")
		publicPath := filepath.Join(gameDir, "res", "values", "public.xml")
		RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrsPath, logger)
	}
}

/**
 * 修复母包中styleable
 */
func RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrPath string, logger models2.LogCallback) {
	logger.LogDebug("开始执行RebuildStyleable")
	publicXml := xml.ParseXml(publicPath)
	attrsXml := xml.ParseXml(attrsPath)
	parseSmali := smali.ParseSmali(styleablePath)
	tag := xml.Tag{
		Name:      "resources",
		Attribute: nil,
		ChildTags: make([]*xml.Tag, 0),
	}

	for _, item := range parseSmali {
		parentTag := xml.Tag{
			Name:      "declare-styleable",
			Attribute: map[string]string{"name": item.Name},
			ChildTags: make([]*xml.Tag, 0),
			Parent:    nil,
		}
		tag.ChildTags = append(tag.ChildTags, &parentTag)
		for k, v := range item.Children {
			findSingleTag := FindSingleTag(publicXml.ChildTags, "public", "id", v)
			attribute := make(map[string]string)
			var tags []*xml.Tag
			if findSingleTag == nil {
				attribute["name"] = strings.ReplaceAll(k, item.Name+"_android_", "android:")
				if strings.HasPrefix(k, "android_lStar") {
					attribute["name"] = strings.ReplaceAll(k, item.Name+"_android_", "")
				} else if strings.HasPrefix(k, "ColorStateListItem_alpha") {
					attribute["name"] = strings.ReplaceAll(k, item.Name+"_", "")
				} else {
					attribute["name"] = strings.ReplaceAll(k, item.Name+"_android_", "android:")
				}

			} else {
				k = findSingleTag.Attribute["name"]
				v = findSingleTag.Attribute["type"]
				if v == "attr" {
					attrTag := FindSingleTag(attrsXml.ChildTags, "attr", "name", k)
					if attrTag.ChildTags != nil && len(attrTag.ChildTags) != 0 {
						for _, item2 := range attrTag.ChildTags {
							attribute2 := make(map[string]string)
							attribute2["name"] = item2.Attribute["name"]
							attribute2["value"] = item2.Attribute["value"]
							tag3 := xml.Tag{
								Name:      item2.Name,
								Attribute: attribute2,
								ChildTags: make([]*xml.Tag, 0),
							}
							tags = append(tags, &tag3)
						}
					} else {
						attribute["format"] = attrTag.Attribute["format"]
					}
				}
				attribute["name"] = k
			}
			parentTag.ChildTags = append(parentTag.ChildTags, &xml.Tag{
				Name:      "attr",
				Attribute: attribute,
				ChildTags: tags,
				Parent:    nil,
			})
		}
	}
	xml.Serializer(tag, xml.XmlHeaderType, newAttrPath)

	err := utils2.Remove(publicPath)
	if err != nil {
		logger.LogDebug("删除", publicPath, err.Error())
	}
}

func RebuildPluginStyleable(pluginPath, pluginName, gamePath string, logger models2.LogCallback) {
	publicPath := filepath.Join(pluginPath, "res", "values", "public.xml")
	attrsPath := filepath.Join(pluginPath, "res", "values", "attrs.xml")
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
	styleablePath := filepath.Join(pluginPath, "smali", packagePath, "R$styleable.smali")
	logger.LogDebug(pluginName, "的styleablePath = ", styleablePath)
	if utils2.Exist(styleablePath) {
		RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrsPath, logger)
	} else {
		err := utils2.Remove(publicPath)
		if err != nil {
			logger.LogDebug("直接删除", publicPath, err.Error())
		}
	}
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
