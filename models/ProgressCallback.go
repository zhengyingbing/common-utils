package models

type ProgressCallback interface {
	Progress(channelId string, num int)
}
