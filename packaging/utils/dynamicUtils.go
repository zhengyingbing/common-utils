package utils

import (
	"bytes"
	"fmt"
	xml "github.com/xyjwsj/xml_parser"
	"github.com/zhengyingbing/common-utils/common/utils"
	"github.com/zhengyingbing/common-utils/packaging/models"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

/**
 * 资源替换
 */
func ReplaceRes(params *models.PreParams, configPath, gameDirPath string, logger models.LogCallback) {
	copyAccessConfig(filepath.Join(configPath, "access.config"), filepath.Join(gameDirPath, "assets", "access.config"))
	replacePackageName(gameDirPath, params.ChannelId, logger)
	replaceIconAndAppName(configPath, gameDirPath, params.ChannelId, logger)
	replaceDynamicConfig(gameDirPath, params.ChannelId, logger)
	replaceGoogleService()
}

func replaceGoogleService() {

}

func replaceDynamicConfig(gameDirPath, channelId string, logger models.LogCallback) {
	cfgPath := filepath.Join(gameDirPath, "assets", "dynamic_config.json")
	if utils.Exist(cfgPath) {
		dynamicConfig := make([]models.DynamicConfig, 0)
		err := utils.ParseToStruct(cfgPath, &dynamicConfig)
		if err != nil {
			logger.LogDebug("parse dynamicConfig failed, reason: " + err.Error())
		}
		models.OperateDynamic(gameDirPath, channelId, dynamicConfig, logger)
	}
}

/**
 * 替换图标和应用名称
 */
func replaceIconAndAppName(configPath, gameDirPath, channelId string, logger models.LogCallback) {
	manifestXml := xml.ParseXml(filepath.Join(gameDirPath, "AndroidManifest.xml"))
	_, tag := FindTag(manifestXml.ChildTags, "application", "")
	app_name := models.GetServerDynamic(channelId)[models.AppName]
	icon_name := models.GetServerDynamic(channelId)[models.IconName]
	iconFolder := ""
	iconName := ""
	appName := ""
	for k, v := range tag.Attribute {
		if k == "android:icon" {
			//@mipmap/icon_app
			if strings.Contains(v, "mipmap") {
				iconFolder = "mipmap-xxhdpi"
			} else {
				iconFolder = "drawable-xxhdpi"
			}
			iconName = v
			//注意加上后缀
			iconName = strings.Split(iconName, "/")[1] + ".png"
			continue
		}
		if k == "android:label" {
			//@string/app_name
			appName = v
			appName = strings.Split(appName, "/")[1]
			continue
		}
	}
	valuesPath := filepath.Join(gameDirPath, "res", "values", "strings.xml")
	valuesXml := xml.ParseXml(valuesPath)

	resourceTag := FindSingleTag(valuesXml.ChildTags, "string", "name", appName)
	resourceTag.Value = fmt.Sprint(app_name)
	xml.Serializer(valuesXml, xml.XmlHeaderType, valuesPath)

	iconPath := filepath.Join(configPath, fmt.Sprint(icon_name))
	if !isPng(iconPath) {
		panic("图标格式错误，非png格式！")
	} else {
		logger.LogDebug("icon check ok !")
	}
	resEntries, _ := os.ReadDir(filepath.Join(gameDirPath, "res"))
	for _, entry := range resEntries {
		if strings.HasPrefix(entry.Name(), "mipmap") || strings.HasPrefix(entry.Name(), "drawable") && entry.IsDir() {
			childs, _ := os.ReadDir(filepath.Join(gameDirPath, "res", entry.Name()))
			for _, child := range childs {
				if child.Name() == iconName {
					utils.Remove(filepath.Join(gameDirPath, "res", entry.Name(), iconName))
				}
			}
		}
	}
	utils.Copy(filepath.Join(configPath, fmt.Sprint(icon_name)), filepath.Join(gameDirPath, "res", iconFolder, iconName), true)
}

func isPng(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	header := make([]byte, 8)
	if _, err = file.Read(header); err != nil {
		return false
	}
	if !bytes.Equal(header, []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return false
	}

	//注意：重置文件指针到开头，以便解码
	if _, err = file.Seek(0, 0); err != nil {
		return false
	}

	//解码整个PNG文件
	_, err = png.Decode(file)
	if err != nil {
		return false
	}
	return true
}

func copyAccessConfig(src string, dst string) {
	utils.ForceCopy(src, dst)
}

/**
 * 替换包名
 */
func replacePackageName(gameDirPath, channelId string, logger models.LogCallback) {
	manifestPath := filepath.Join(gameDirPath, "AndroidManifest.xml")
	manifestXml := xml.ParseXml(manifestPath)
	gamePackage := manifestXml.Attribute["package"]
	logger.LogDebug("包名：", gamePackage)
	pkgName := models.GetServerDynamic(channelId)[models.BundleId]
	utils.ReplaceFile(manifestPath, gamePackage, fmt.Sprint(pkgName))
	utils.ReplaceFile(manifestPath, "hlApplicationId", fmt.Sprint(pkgName))

}
