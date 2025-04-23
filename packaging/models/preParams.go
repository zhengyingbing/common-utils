package models

type PreParams struct {
	JavaHome, AndroidHome, BuildPath, Channel, ChannelId, HomePath, ExpandPath, GamePath, KeystoreName string
	Plugins                                                                                            []string
}
