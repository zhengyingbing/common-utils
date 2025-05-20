package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

/**
 * 日志系统
 * 军工级可靠性
 * 航母级吞吐量
 * 太空级可观测性
 * 磁悬浮级性能
 */

var logger ChannelLogger
var bf *bytes.Buffer = bytes.NewBuffer(nil)

type ChannelLogger struct {
	file   *os.File
	Buffer *bytes.Buffer //内存缓冲
	mu     sync.Mutex
}

//type LogSystem struct {
//}

func LogSystem() ChannelLogger {
	return ChannelLogger{}
}

func (ChannelLogger) LogVerbose(data ...any) {
	logStr := fmt.Sprintln(data)
	writeLog(logStr)
	log.Println(data)
}

func (ChannelLogger) LogDebug(data ...any) {
	logStr := fmt.Sprintln(data)
	writeLog(logStr)
	log.Println(data)
}

func (ChannelLogger) LogInfo(data ...any) {
	logStr := fmt.Sprintln(data)
	writeLog(logStr)
	//log.Println(data)
	log.Println(append([]interface{}{"[INFO] - ", time.DateTime}, data...)...)
}

func Init(channelId, logDir string) ChannelLogger {
	logPath := filepath.Join(logDir, "logs", strings.Join([]string{channelId, "_", time.Now().Format("20060102-15040511"), ".log"}, ""))
	os.MkdirAll(logDir, 0755)
	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logger = ChannelLogger{
		file:   f,
		Buffer: bf,
	}
	StartFlusher(3 * time.Second)

	return logger
}

func Write(data ...any) {
	log.Println(append([]interface{}{"[INFO] - ", time.DateTime}, data...)...)
	logStr := fmt.Sprintln(data)
	bf.WriteString(time.Now().Format("2006-01-02 15:04:05") + logStr + "\n")
}

func writeLog(msg string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.Buffer.WriteString(time.Now().Format("2006-01-02 15:04:05") + msg + "\n")
	if logger.Buffer.Len() > 10<<10 {
		flushToDisk()
	}
}

func StartFlusher(val time.Duration) {
	go func() {
		ticker := time.NewTicker(val)
		for range ticker.C {
			logger.mu.Lock()
			flushToDisk()
			logger.mu.Unlock()
		}
	}()
}

func flushToDisk() {
	if logger.file == nil && logger.Buffer.Len() == 0 {
		return
	}
	logger.file.Write(logger.Buffer.Bytes())
	logger.Buffer.Reset() //清空缓冲区
}

// 崩溃时刷盘
func Shutdown() {
	flushToDisk()
	logger.file.Sync() //系统级sync
}
