package models

type PreParams struct {
	JavaHome, AndroidHome, ApkPath, RootPath, BuildPath, OutPutPath                    string
	PackageName, ProductId, ProductName, ApkName, ChannelName, ChannelId, KeystoreName string
	Plugins                                                                            []string
}
