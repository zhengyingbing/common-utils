package packaging

import xml "github.com/xyjwsj/xml_parser"

// 插件apk/res合并到母包/res中
func MergeRes(src, dst string, isForced bool) error {

	return nil
}

func RebuildStyleable(styleablePath, publicPath, attrsPath, newAttrPath string) {

	publicXml := xml.ParseXml(publicPath)
	attrsXml := xml.ParseXml(attrsPath)

}
