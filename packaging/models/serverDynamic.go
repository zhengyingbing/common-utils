package models

type OperateType string

const (
	CopyClass OperateType = "copyClass"
	Copy      OperateType = "copy"
	Replace   OperateType = "replace"
	Delete    OperateType = "delete"
	Move      OperateType = "move"
)

type DynamicConfig struct {
	ContentPath string    `json:"contentPath"`
	Operates    []Operate `json:"operates"`
}

type Operate struct {
	Operate OperateType              `json:"operate"`
	Config  map[string](interface{}) `json:"config"`
}

type ReplaceConfig struct {
	ConstantVal string `json:"constantVal"`
	Server      bool   `json:"server"`
	TarName     string `json:"tarName"`
	VarName     string `json:"varName"`
}

type MoveConfig struct {
	Server bool   `json:"server"`
	Source string `json:"source"`
	Target string `json:"target"`
}
