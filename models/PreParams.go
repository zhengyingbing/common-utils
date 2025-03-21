package models

type PreParams struct {
	JavaHome, AndroidHome, BuildPath, Channel, ChannelId, ExpandPath, GamePath string
	Plugins                                                                    []string
}
