package smali

import (
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"strings"
)

func ParseSmali(path string) []HexSmaliField {
	smaliMgrInstance := CreateSmaliMgrInstance()
	registerMgrInstance := CreateRegisterMgrInstance()
	grammarInstance := CreateGrammarInstance(smaliMgrInstance, registerMgrInstance)

	pmAddr := false
	name := ""
	utils.ReadLine(path, func(err error, line int, content string) bool {
		if content == "" {
			return false
		}
		content = strings.TrimSpace(content)

		if strings.HasPrefix(content, ":array_") {
			name = strings.ReplaceAll(content, ":", "")
			return false
		}

		if content == ".end array-data" {
			return false
		}
		if pmAddr {
			ref := registerMgrInstance.getRef(name)
			if ref == nil {
				ints := make([]int, 0)
				ref = append(ints, utils.Hex2Dec(content))
			} else {
				if val, ok := ref.([]int); ok {
					ref = append(val, utils.Hex2Dec(content))
				}
			}
			registerMgrInstance.saveRef(name, ref)
		}
		return false
	})

	return smaliMgrInstance.AllDexFields()
}
