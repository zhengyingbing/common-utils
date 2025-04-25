//go:build test

// test.go
package main

import (
	"os"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/utils"
)

const (
	macSplit     = "/"
	windowsSplit = "\\"
)

func main() {
	gamePath := "C:\\Users\\zheng\\Desktop\\9\\gameDir"
	//pPath := "C:\\Users\\zheng\\Desktop\\9\\hoolaiDir"
	utils.MergeSmaliFiles(gamePath)
	utils.GameRepairStyleable(gamePath, models.LogImpl{})
	//utils.RebuildPluginStyleable(pPath, "hoolai", gamePath, models.LogImpl{})
	//remove2()
	//newAttrsPath := "C:\\Users\\zheng\\Desktop\\9\\values_attrs.xml"
	//
	//
	//styleablePath := "C:\\Users\\zheng\\Desktop\\9\\R$styleable.smali"
	//publicPath := "C:\\Users\\zheng\\Desktop\\9\\public.xml"
	//attrsPath := "C:\\Users\\zheng\\Desktop\\9\\attrs.xml"
	////attrsPath := "C:\\apktool\\home\\gameDir\\res\\values\\attrs.xml"
	//publicXml := xml.ParseXml(publicPath)
	//attrsXml := xml.ParseXml(attrsPath)
	//parseSmali := smali.ParseSmali(styleablePath)
	//tag := xml.Tag{
	//	Name:      "resources",
	//	Attribute: nil,
	//	ChildTags: make([]*xml.Tag, 0),
	//}
	//
	//for _, item := range parseSmali {
	//	parentTag := xml.Tag{
	//		Name:      "declare-styleable",
	//		Attribute: map[string]string{"name": item.Name},
	//		ChildTags: make([]*xml.Tag, 0),
	//		Parent:    nil,
	//	}
	//	tag.ChildTags = append(tag.ChildTags, &parentTag)
	//	for k, v := range item.Children {
	//		findSingleTag := utils.FindSingleTag(publicXml.ChildTags, "public", "id", v)
	//		attribute := make(map[string]string)
	//		childTag := make([]*xml.Tag, 0)
	//		if findSingleTag == nil {
	//			attribute["name"] = strings.ReplaceAll(k, item.Name+"_android_", "android:")
	//			if strings.HasPrefix(k, "android_lStar") {
	//				attribute["name"] = strings.ReplaceAll(k, item.Name+"_android_", "")
	//			} else if strings.HasPrefix(k, "ColorStateListItem_alpha") {
	//				attribute["name"] = strings.ReplaceAll(k, item.Name+"_", "")
	//			} else {
	//				attribute["name"] = strings.ReplaceAll(k, item.Name+"_android_", "android:")
	//			}
	//
	//		} else {
	//			k = findSingleTag.Attribute["name"]
	//			v = findSingleTag.Attribute["type"]
	//			if v == "attr" {
	//				attrTag := utils.FindSingleTag(attrsXml.ChildTags, "attr", "name", k)
	//				if attrTag.ChildTags != nil && len(attrTag.ChildTags) != 0 {
	//
	//					for _, item2 := range attrTag.ChildTags {
	//						attribute2 := make(map[string]string)
	//						attribute2["name"] = item2.Attribute["name"]
	//						attribute2["value"] = item2.Attribute["value"]
	//						tag3 := xml.Tag{
	//							Name:      item2.Name,
	//							Attribute: attribute2,
	//							ChildTags: make([]*xml.Tag, 0),
	//						}
	//						childTag = append(childTag, &tag3)
	//					}
	//				} else {
	//					attribute["format"] = attrTag.Attribute["format"]
	//				}
	//			}
	//			attribute["name"] = k
	//
	//		}
	//		parentTag.ChildTags = append(parentTag.ChildTags, &xml.Tag{
	//			Name:      "attr",
	//			Attribute: attribute,
	//			ChildTags: childTag,
	//			Parent:    nil,
	//		})
	//	}
	//}
	//xml.Serializer(tag, xml.XmlHeaderType, newAttrsPath)

	//gamePath := "C:\\apktool\\home\\3015_10302"
	//preParams := models2.PreParams{
	//	Channel:      "douyin",
	//	ChannelId:    "10302",
	//	GamePath:     gamePath,
	//	KeystoreName: "aygd.keystore",
	//}
	//cfg := make(map[string]string)
	//cfg[models2.AppName] = "douyin" + "Demo"
	//cfg[models2.IconName] = "ic_launcher.png"
	//cfg[models2.TargetSdkVersion] = "30"
	//cfg[models2.DexMethodCounters] = "60000"
	//cfg[models2.BundleId] = "com.hoolai.sf3.bytedance.gamecenter"
	//cfg[models2.SignVersion] = "2"
	//cfg[models2.KeystoreAlias] = "aygd3"
	//cfg[models2.KeystorePass] = "aygd3123"
	//cfg[models2.KeyPass] = "aygd3123"
	//cfg["appId"] = "614371"
	//models2.SetServerDynamic("10302", cfg)
	//jarsigner := filepath.Join("C:\\apktool\\resources\\java", "win", "jre", "bin", "jarsigner.exe")
	//apksigner := filepath.Join("C:\\apktool\\resources\\android", "windows", "apksigner.bat")
	//zipalign := filepath.Join("C:\\apktool\\resources\\android", "windows", "zipalign.exe")
	//utils2.SignApk(gamePath, jarsigner, apksigner, zipalign, &preParams, &models2.LogImpl{})
}

type LoginCallback interface {
	OnSuccess(uid, token string)
	onFailed(err string)
}

type HandleLogin struct{}

func (h HandleLogin) OnSuccess(uid, token string) {

}

func remove2() error {
	src := "C:\\apktool\\home\\1_1"
	dst := "C:\\apktool\\tt3"
	err := os.Rename(src, dst)
	go func() {
		err = utils2.Remove(dst)
	}()
	if err != nil {
		println("删除失败", err.Error())
	}

	return nil
}
