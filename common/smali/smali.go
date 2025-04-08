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
			pmAddr = false
			return false
		}

		if strings.HasPrefix(content, ".array-data") {
			pmAddr = true
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

	pMethod := false

	utils.ReadLine(path, func(err error, line int, content string) bool {
		if content == "" {
			return false
		}

		content = strings.TrimSpace(content)
		if strings.HasPrefix(content, ".end method") {
			pMethod = false
		}

		if strings.HasPrefix(content, ".method") {
			pMethod = true
			return false
		}

		if pMethod {
			if parseMethod(content, grammarInstance) {
				pMethod = false
			}
			return false
		}

		if strings.HasPrefix(content, ".field") {
			parseField(content, smaliMgrInstance)
			return false
		}
		return false
	})
	return smaliMgrInstance.AllDexFields()
}

func parseMethod(line string, grammar *Grammar) bool {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "return-") {
		return true
	}
	if strings.HasPrefix(line, ".locals") {
		return false
	}
	index := strings.Index(line, " ")
	key := line[0:index]
	params := line[index+1:]
	params = strings.ReplaceAll(params, " ", "")
	split := strings.Split(params, ",")
	grammar.GrammarCode(key, split)
	return false
}

func parseField(line string, mgr *SmaliMgr) {
	var replace string
	if strings.Contains(line, ".field public static") {
		replace = ".field public static "
	}
	if strings.Contains(line, ".field public static final") {
		replace = ".field public static final "
	}
	line = strings.ReplaceAll(line, replace, "")
	line = strings.TrimSpace(line)
	if strings.Contains(line, "=") && !strings.Contains(line, "=null") {
		spilt := strings.Split(line, "=")
		fieldName := strings.Split(spilt[0], ":")[0]
		val := utils.Hex2Dec(spilt[1])
		mgr.AddChildField(fieldName, val)
	} else {
		fieldName := strings.Split(line, ":")[0]
		mgr.AddField(fieldName)
	}
}
