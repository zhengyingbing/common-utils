package models

type LogCallback interface {
	Println(data ...any)
	Printf(str string, data ...any)
}
