package utils

import (
	xml "github.com/xyjwsj/xml_parser"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/smali"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strings"
)

// 插件apk/res合并到母包/res中
func MergeRes(src, pluginName, dst string, isForced bool) error {
	RebuildPluginStyleable(src, pluginName, dst)
	src = filepath.Join(src, "res")
	dst = filepath.Join(dst, "res")
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "values") {
			err = mergeValues(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()), isForced)
		} else {
			err = utils.Move(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()), isForced)
		}
	}
	return err
}

func mergeValues(src string, dst string, isForced bool) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		mergeXml(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()), isForced)
	}
	return nil
}

func mergeXml(src string, dst string, isForced bool) {
	if !utils.Exist(dst) {
		_ = utils.Move(src, dst, true)
		return
	}
	srcXml := xml.ParseXml(src)
	dstXml := xml.ParseXml(dst)
	uniqueMap := make(map[string]bool)

	for _, tag := range dstXml.ChildTags {
		uniqueMap[tag.Attribute["name"]] = true
	}

	for _, tag := range srcXml.ChildTags {
		if !uniqueMap[tag.Attribute["name"]] || isForced {
			dstXml.ChildTags = append(dstXml.ChildTags, tag)
			uniqueMap[tag.Attribute["name"]] = true
		}
	}
	xml.Serializer(dstXml, xml.XmlHeaderType, dst)
	uniqueMap = nil
}

func mergeOtherRes(src, dst string, isForced bool) error {
	if utils.Exist(dst) {
		entries, err := os.ReadDir(src)
		for _, entry := range entries {
			_ = utils.Move(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name()), isForced)
		}
		return err
	} else {
		return utils.Move(src, dst, isForced)
	}
}

func RebuildPluginStyleable(pluginPath, pluginName, gamePath string) {
	publicPath := filepath.Join(pluginPath, "res", "values", "public.xml")
	attrsPath := filepath.Join(pluginPath, "res", "values", "attrs.xml")
	newAttrsPath := filepath.Join(pluginPath, "res", "values", "values_attrs.xml")
	if utils.Exist(newAttrsPath) || strings.EqualFold(pluginName, "firebase") {
		err := os.Remove(publicPath)
		if err != nil {
			return
		}
		return
	}
	packageName := PackageName(filepath.Join(pluginPath, "AndroidManifest.xml"))
	packagePath := strings.Replace(packageName, ".", utils.Symbol(), -1)
	styleablePath := filepath.Join(pluginPath, "smali", packagePath, "R$styleable.smali")
	println(pluginName, "的styleablePath = ", styleablePath)
	if utils.Exist(styleablePath) {
		RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrsPath)
	} else {
		err := utils.Remove(publicPath)
		if err != nil {
			println("直接删除", publicPath, err.Error())
		}
	}
}

/**
 * 修复母包中styleable
 */
func RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrPath string) {
	println("开始执行RebuildStyleable")
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

	err := utils.Remove(publicPath)
	if err != nil {
		println("删除", publicPath, err.Error())
	}
}

func BuildRes(aapt2, gamePath string, logger models.LogCallback) {
	logger.Println("开始构建res")
	shellString := aapt2 + " compile --dir " + filepath.Join(gamePath, "res") + " -o " + filepath.Join(gamePath, "res.zip")
	err := utils.ExecuteShell(shellString)
	if err != nil {
		logger.Println("res.zip build failed!")
		panic(err.Error())
	} else {
		logger.Println("res.zip build success!")
	}
}
