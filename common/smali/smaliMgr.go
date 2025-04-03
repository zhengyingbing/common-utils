package smali

import "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"

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
	return fields
}
