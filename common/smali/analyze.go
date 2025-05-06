package smali

import (
	"github.com/zhengyingbing/common-utils/common/utils"
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
		grammar.aPut(params)
	}
}

func (grammar *Grammar) constKey(params []string) {
	//const/16 v0, 0x1d
	//params[0] v0
	//params[1] 0x1d
	grammar.registerMgr.saveLocal(params[0], utils.Hex2Dec(params[1]))
}

func (grammar *Grammar) newArray(params []string) {
	//new-array v0, v0, [I
	//params[0] v0
	//params[1] v0
	arr := make([]int, grammar.registerMgr.getLocal(params[1]))
	grammar.registerMgr.saveRef(params[0], arr)
}

func (grammar *Grammar) fillArrayData(params []string) {
	//fill-array-data v0, :array_0
	//params[0] v0
	//params[1][1:] array_0
	ref := grammar.registerMgr.getRef(params[0])
	if _, ok := ref.([]int); ok {
		adds := grammar.registerMgr.getRef(params[1][1:])
		if addr, ok := adds.([]int); ok {
			grammar.registerMgr.saveRef(params[0], addr)
		}
	}
}

func (grammar *Grammar) sPutObject(params []string) {
	//sput-object v0, Lcom/hoolai/demo/R$styleable;->ActionBar:[I
	split := strings.Split(params[1], "->")
	name := strings.Split(split[1], ":")[0]
	ref := grammar.registerMgr.getRef(params[0])
	//name = ActionBar
	//ref = sput-object v0
	if val, ok := ref.([]int); ok {
		grammar.smaliMgr.ConfigFieldVal(name, val)
	}
}

func (grammar *Grammar) aPut(params []string) {
	//aput v2, v1, v3
	//params[1] = v1
	ref := grammar.registerMgr.getRef(params[1])
	//ref是任意类型，尝试将其转换为[]int类型
	if val, ok := ref.([]int); ok {
		index := grammar.registerMgr.getLocal(params[2])
		//修改切片元素
		val[index] = grammar.registerMgr.getLocal(params[0])
	}
}
