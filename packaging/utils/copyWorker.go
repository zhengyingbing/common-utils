package utils

import (
	"github.com/zhengyingbing/common-utils/common/utils"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CopyTask struct {
	//ChannelId string
	SrcPath string
	DstPath string
}

func copyWorker(queue chan CopyTask, s *sync.WaitGroup) {
	defer s.Done()
	for task := range queue {
		start := time.Now()
		copyPlugin(task)
		log.Printf("拷贝完成 %s %v", task.DstPath, time.Since(start))
	}
}

func copyPlugin(task CopyTask) {
	dst0 := filepath.Join(task.DstPath, filepath.Base(task.SrcPath))
	os.MkdirAll(dst0, 0755)
	log.Println(dst0)
	log.Printf("开始拷贝 %s %s", task.SrcPath, dst0)
	src, _ := os.Open(task.SrcPath)
	defer src.Close()
	dst, _ := os.Create(dst0)
	defer dst.Close()
	utils.Copy(task.SrcPath, dst0, true)

}
