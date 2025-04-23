package smali

import (
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"strings"
)

type HexSmaliField struct {
	Name     string            `json:"name`
	Children map[string]string `json:"children"`
}

type SmaliField struct {
	Name     string         `json:"name"`
	Chileren map[string]int `json:"children"`
}

type SmaliMgr struct {
	fields []*SmaliField
}

func CreateSmaliMgrInstance() *SmaliMgr {
	return &SmaliMgr{
		fields: make([]*SmaliField, 0),
	}
}

func (smaliMgr *SmaliMgr) AllDexFields() []HexSmaliField {
	fields := make([]HexSmaliField, 0)
	for _, item := range smaliMgr.fields {
		cMap := make(map[string]string)
		for k, v := range item.Chileren {
			cMap[k] = utils.Dec2Hex(v)
		}
		fields = append(fields, HexSmaliField{
			Name:     item.Name,
			Children: cMap,
		})
	}
	//print("AllDexFields的长度：", len(fields))
	return fields
}

func (smaliMgr *SmaliMgr) ConfigFieldVal(name string, valArr []int) {
	for _, item := range smaliMgr.fields {
		if item.Name == name {
			for k, v := range item.Chileren {
				item.Chileren[k] = valArr[v]
			}
		}
	}
}

func (smaliMgr *SmaliMgr) AddField(name string) {
	exist := false
	for _, item := range smaliMgr.fields {
		if item.Name == name {
			exist = true
			break
		}
	}
	if !exist {
		smaliMgr.fields = append(smaliMgr.fields, &SmaliField{
			Name:     name,
			Chileren: make(map[string]int),
		})
	}
}

func (smaliMgr *SmaliMgr) AddChildField(name string, val int) {
	for i := len(smaliMgr.fields) - 1; i >= 0; i-- {
		if strings.HasPrefix(name, smaliMgr.fields[i].Name+"_") {
			smaliMgr.fields[i].Chileren[name] = val
			break
		}
	}
}
