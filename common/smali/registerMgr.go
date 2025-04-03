package smali

type RegisterMgr struct {
	//本地寄存器
	local map[string]int
	//引用寄存器
	ref map[string]any
}

func CreateRegisterMgrInstance() *RegisterMgr {
	return &RegisterMgr{
		local: make(map[string]int, 10),
		ref:   make(map[string]any, 10),
	}
}

func (registerMgr *RegisterMgr) saveRef(key string, val any) {
	registerMgr.ref[key] = val
}

func (registerMgr *RegisterMgr) getRef(key string) any {
	for k, v := range registerMgr.ref {
		if k == key {
			return v
		}
	}
	return nil
}
