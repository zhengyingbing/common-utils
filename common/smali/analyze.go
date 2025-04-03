package smali

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
