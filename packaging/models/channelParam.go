package models

type ChannelParam struct {
	AppName     string
	ChannelId   string
	ChannelName string
	ChannelDesc string
	Version     string
	//打包状态：未开始1，打包中2，打包成功0，打包失败-1
	status      int
	Progress    int
	IsChecked   bool
	PackageName string
}
