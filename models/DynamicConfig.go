package models

var dynamicCfgMap map[string]map[string]any

const (
	AppName          = "appName"
	IconName         = "iconName"
	BundleId         = "bundleId"
	KeystoreAlias    = "keystoreAlias"
	KeystorePass     = "keystorePass"
	KeyPass          = "keyPass"
	TargetSdkVersion = "targetSdkVersion"
)

func SetChannelDynamicConfig(channelId string, cfg map[string]any) {
	if dynamicCfgMap == nil {
		dynamicCfgMap = make(map[string]map[string]any)
	}
	dynamicCfgMap[channelId] = cfg
}

func GetDynamicConfig(channelId string) map[string]any {
	return dynamicCfgMap[channelId]
}
