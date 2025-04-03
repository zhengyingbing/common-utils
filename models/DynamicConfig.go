package models

var dynamicCfgMap map[string]map[string]any

func SetChannelDynamicConfig(channelId string, cfg map[string]any) {
	if dynamicCfgMap == nil {
		dynamicCfgMap = make(map[string]map[string]any)
	}
	dynamicCfgMap[channelId] = cfg
}

func GetDynamicConfig(channelId string) map[string]any {
	return dynamicCfgMap[channelId]
}
