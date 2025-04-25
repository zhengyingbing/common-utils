package utils

import (
	xml "github.com/xyjwsj/xml_parser"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strings"
)

// 插件apk/res合并到母包/res中
func MergeRes(src, pluginName, dst string, isForced bool, logger models.LogCallback) error {
	RebuildPluginStyleable(src, pluginName, dst, logger)
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
