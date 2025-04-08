package packaging

import (
	xml "github.com/xyjwsj/xml_parser"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/smali"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"strings"
)

// 插件apk/res合并到母包/res中
func MergeRes(src, pluginName, dst string, isForced bool) error {
	RebuildPluginStyleable(src, pluginName, dst)
	return nil
}

func RebuildPluginStyleable(pluginPath, pluginName, gamePath string) {
	publicXml := filepath.Join(pluginPath, "res", "values", "public.xml")
	attrsXml := filepath.Join(pluginPath, "res", "values", "attrs.xml")
	newAttrsXml := filepath.Join(pluginPath, "res", "values", "values_attrs.xml")
	if utils.Exist(newAttrsXml) || strings.EqualFold(pluginName, "firebase") {
		err := os.Remove(publicXml)
		if err != nil {
			return
		}
		return
	}
	if utils.Exist(publicXml) {
		packageName := PackageName(filepath.Join(pluginPath, "AndroidManifest.xml"))
		packagePath := strings.Replace(packageName, ".", utils.Symbol(), -1)
		styleablePath := filepath.Join(gamePath, "smali", packagePath, "R$styleable.smali")
		println(pluginName, "的styleablePath = ", styleablePath)
		if utils.Exist(styleablePath) {
			RebuildStyleable(styleablePath, publicXml, attrsXml, newAttrsXml)
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
}
