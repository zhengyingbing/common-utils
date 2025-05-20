package models

type LogCallback interface {
	LogInfo(data ...any)
	LogDebug(data ...any)
	LogVerbose(data ...any)
	//Printf(str string, data ...any)
}
