package smali

import (
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"strings"
)

type Grammar struct {
	smaliMgr    *SmaliMgr
	registerMgr *RegisterMgr
}

func CreateGrammarInstance(smaliMgr *SmaliMgr, registerMgr *RegisterMgr) *Grammar {
	return &Grammar{
		registerMgr: registerMgr,
		smaliMgr:    smaliMgr,
	}
}

func (grammar *Grammar) GrammarCode(key string, params []string) {
	switch {
	case strings.HasPrefix(key, "const"):
		grammar.constKey(params)
	case key == "new-array":
		grammar.newArray(params)
	case key == "fill-array-data":
		grammar.fillArrayData(params)
	case key == "sput-object":
		grammar.sPutObject(params)
	case key == "aput":
		grammar.newArray(params)
	}
}

func (grammar *Grammar) constKey(params []string) {
	grammar.registerMgr.saveLocal(params[0], utils.Hex2Dec(params[1]))
}

func (grammar *Grammar) newArray(params []string) {
	arr := make([]int, grammar.registerMgr.getLocal(params[1]))
	grammar.registerMgr.saveRef(params[0], arr)
}

func (grammar *Grammar) fillArrayData(params []string) {
	ref := grammar.registerMgr.getRef(params[0])
	if _, ok := ref.([]int); ok {
		adds := grammar.registerMgr.getRef(params[1][1:])
		if addr, ok := adds.([]int); ok {
			grammar.registerMgr.saveRef(params[0], addr)
		}
	}
}

func (grammar *Grammar) sPutObject(params []string) {
	split := strings.Split(params[1], "->")
	name := strings.Split(split[1], ":")[0]
	ref := grammar.registerMgr.getRef(params[0])
	if val, ok := ref.([]int); ok {
		grammar.smaliMgr.ConfigFieldVal(name, val)
	}
}

func (grammar *Grammar) aPut(params []string) {
	ref := grammar.registerMgr.getRef(params[1])
	if val, ok := ref.([]int); ok {
		index := grammar.registerMgr.getLocal(params[2])
		val[index] = grammar.registerMgr.getLocal(params[0])
	}
}
