package utils

import (
	"fmt"
	utils2 "github.com/zhengyingbing/common-utils/common/utils"
	"github.com/zhengyingbing/common-utils/packaging/models"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

/**
 * 插件dir路径， 多渠道文件夹目录
 */
func CopyPlugins(srcDirs []string, buildPaths []string, progress models.ProgressCallback) error {
	sem := make(chan struct{}, runtime.NumCPU()*2) // 信号量控制并发度
	errChan := make(chan error, len(buildPaths)*len(srcDirs))
	//taskQueue := make(chan CopyTask, len(buildPath)*len(srcDirs))
	var wg sync.WaitGroup
	//for i := 0; i < runtime.NumCPU()*2; i++ {
	//	wg.Add(1)
	//	go copyWorker(taskQueue, &wg)
	//}
	for _, channel := range buildPaths {
		if err := os.MkdirAll(channel, 0755); err != nil {
			return err
		}

		for _, srcFile := range srcDirs {
			wg.Add(1)
			go func(src, dst string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() {}()

				dstPath := filepath.Join(dst, filepath.Base(src))
				if err := atomicCopy(src, dstPath); err != nil {
					errChan <- fmt.Errorf("拷贝失败 %s→%s: %v", src, dstPath, err)
				}
			}(srcFile, channel)

		}
	}
	wg.Wait()
	return nil
}

// 原子化拷贝操作
func atomicCopy(src, dst string) error {
	// 1. 先拷贝到临时文件
	tmpFile := dst + ".tmp"
	if err := utils2.Copy(src, tmpFile, true); err != nil {
		return err
	}

	// 2. 原子性重命名
	return os.Rename(tmpFile, dst)
}

type CopyTask struct {
	//ChannelId string
	SrcPath string
	DstPath string
}

func copyWorker(queue chan CopyTask, s *sync.WaitGroup) {
	defer s.Done()
	for task := range queue {
		start := time.Now()
		CopyPlugin(task)
		Write("拷贝完成 %s %v", task.DstPath, time.Since(start))
	}
}

func CopyPlugin(task CopyTask) {
	dst0 := filepath.Join(task.DstPath, filepath.Base(task.SrcPath))
	os.MkdirAll(dst0, 0755)
	Write("开始拷贝 %s %s", task.SrcPath, dst0)
	src, _ := os.Open(task.SrcPath)
	defer src.Close()
	dst, _ := os.Create(dst0)
	defer dst.Close()
	utils2.Copy(task.SrcPath, dst0, true)

}
