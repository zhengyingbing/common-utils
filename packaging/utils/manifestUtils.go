package utils

import (
	xml "github.com/xyjwsj/xml_parser"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strings"
)

func MergeManifest(src, dst string, logger models.LogCallback) error {
	srcXml := xml.ParseXml(src)
	dstXml := xml.ParseXml(dst)
	mergePermission(&srcXml, &dstXml)
	mergeQueries(&srcXml, &dstXml)
	mergeApplication(&srcXml, &dstXml)

	xml.Serializer(dstXml, xml.XmlHeaderType, dst)

	return nil
}

func mergePermission(srcXml *xml.Tag, dstXml *xml.Tag) {
	uniqueMap := make(map[*xml.Tag]bool)
	for _, tag := range dstXml.ChildTags {
		if !(strings.Contains(tag.Name, "queries") || strings.Contains(tag.Name, "application")) {
			uniqueMap[tag] = true
		}
	}
	for _, tag := range srcXml.ChildTags {
		if !(strings.Contains(tag.Name, "queries") || strings.Contains(tag.Name, "application")) {
			if !uniqueMap[tag] {
				dstXml.ChildTags = append(dstXml.ChildTags, tag)
				uniqueMap[tag] = true
			}
		}
	}
	uniqueMap = nil
}

func mergeQueries(srcXml *xml.Tag, dstXml *xml.Tag) {
	_, srcApplication := FindTag(srcXml.ChildTags, "queries", "")
	_, dstApplication := FindTag(dstXml.ChildTags, "queries", "")
	if srcApplication != nil {
		dstApplication.ChildTags = append(dstApplication.ChildTags, srcApplication.ChildTags...)
	}

}

func mergeApplication(srcXml *xml.Tag, dstXml *xml.Tag) {
	_, srcApplication := FindTag(srcXml.ChildTags, "application", "")
	_, dstApplication := FindTag(dstXml.ChildTags, "application", "")
	dstApplication.ChildTags = append(dstApplication.ChildTags, srcApplication.ChildTags...)
}

func FindTag(tags []*xml.Tag, tagName, androidName string) (int, *xml.Tag) {
	for index, item := range tags {
		if item.Name == tagName {
			if androidName == "" {
				return index, item
			} else {
				if item.Attribute != nil && androidName == item.Attribute["android:name"] {
					return index, item
				}
			}
		}
	}
	return 0, nil
}

func FindSingleTag(tags []*xml.Tag, tagName, attrName, attrVal string) *xml.Tag {
	for _, item := range tags {
		if item.Name == tagName {
			if val, ok := item.Attribute[attrName]; ok && val == attrVal {
				return item
			}
		}
	}
	return nil
}

func PackageName(manifestPath string) string {
	gameXml := xml.ParseXml(manifestPath)
	packageName := ""
	for key, value := range gameXml.Attribute {
		if strings.HasPrefix(key, "package") {
			packageName = value
		}
	}
	return packageName
}
