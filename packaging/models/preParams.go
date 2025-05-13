package models

type PreParams struct {
	JavaHome, AndroidHome, ApkPath, RootPath, OutPutPath                  string
	PackageName, ProductId, ApkName, ChannelName, ChannelId, KeystoreName string
	Plugins                                                               []string
}
