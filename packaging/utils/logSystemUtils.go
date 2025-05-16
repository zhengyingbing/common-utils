package utils

import (
	"bytes"
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

type ChannelLogger struct {
	file   *os.File
	buffer *bytes.Buffer //内存缓冲
	mu     sync.Mutex
}

type LogManager struct {
	loggers map[string]*ChannelLogger
	baseDir string
	mu      sync.RWMutex
}

// 初始化
func NewLogManager(basePath string) *LogManager {
	return &LogManager{
		loggers: make(map[string]*ChannelLogger),
		baseDir: basePath,
	}
}

func (lm *LogManager) GetLogger(channelId string) *ChannelLogger {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	if logger, exists := lm.loggers[channelId]; exists {
		return logger
	}
	logPath := filepath.Join(lm.baseDir, strings.Join([]string{channelId, "_", time.Now().Format("20060102-15040511"), ".log"}, ""))
	os.MkdirAll(lm.baseDir, 0755)
	f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logger := &ChannelLogger{
		file:   f,
		buffer: bytes.NewBuffer(nil),
	}
	lm.loggers[channelId] = logger
	return logger
}
func (logger *ChannelLogger) WriteLog(msg string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	logger.buffer.WriteString(time.Now().Format("2006-01-02 15:04:05 ") + msg + "\n")
	if logger.buffer.Len() > 10<<10 {
		logger.flushToDisk()
	}
}

func (lm *LogManager) StartFlusher(val time.Duration) {
	go func() {
		ticker := time.NewTicker(val)
		for range ticker.C {
			lm.mu.RLock()
			for _, logger := range lm.loggers {
				logger.mu.Lock()
				logger.flushToDisk()
				logger.mu.Unlock()
			}
			lm.mu.RUnlock()
		}
	}()
}

func (logger *ChannelLogger) flushToDisk() {
	if logger.buffer.Len() == 0 {
		return
	}
	logger.file.Write(logger.buffer.Bytes())
	logger.buffer.Reset() //清空缓冲区
}

// 崩溃时刷盘
func (lm *LogManager) Shutdown() {
	for _, cl := range lm.loggers {
		cl.flushToDisk()
		cl.file.Sync() //系统级sync
	}
}
