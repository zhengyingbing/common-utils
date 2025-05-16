package models

type Progress struct {
	Callback func(channelId string, num int) `json:"-"`
}
