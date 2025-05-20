// test.go
package main

import (
	"fmt"
	"github.com/zhengyingbing/common-utils/common/utils"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	macSplit     = "/"
	windowsSplit = "\\"
)

func main() {

	//productParam := []string{"1", "2", "channelId", "channelName", "4", "5"}
	//apkName := strings.Join(productParam, "_")
	//apkName = strings.Replace(apkName, "channelId", "10010", -1)
	//apkName = strings.Replace(apkName, "channelName", "hoolai", -1) + ".apk"
	//println("名字：" + apkName)
	//p := "C:\\apktool\\gameDir"
	//utils.Remove("C:\\apktool\\build\\2")
	apks := []string{"C:\\apktool\\gameDir", "C:\\apktool\\sdk\\expand\\coreDir", "C:\\apktool\\sdk\\expand\\fastsdkDir", "C:\\apktool\\sdk\\expand\\hoolaiDir", "C:\\apktool\\sdk\\expand\\xiaomiDir"}
	//targetDirs := map[string]string{"1": "C:\\apktool\\build\\1", "2": "C:\\apktool\\build\\2", "3": "C:\\apktool\\build\\3"}
	channels := []string{"C:\\apktool\\build\\1", "C:\\apktool\\build\\2", "C:\\apktool\\build\\3"}
	utils.Remove("C:\\apktool\\build\\1")
	utils.Remove("C:\\apktool\\build\\2")
	utils.Remove("C:\\apktool\\build\\3")
	os.MkdirAll("C:\\apktool\\build\\1", 0755)
	os.MkdirAll("C:\\apktool\\build\\2", 0755)
	os.MkdirAll("C:\\apktool\\build\\3", 0755)
	CopyPlugins(channels, apks)

}
func CopyPlugins(channels []string, apks []string) error {
	taskQueue := make(chan CopyTask, len(channels)*len(apks))

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU()*2; i++ {
		wg.Add(1)
		go copyWorker(taskQueue, &wg)
	}
	for _, channel := range channels {
		for _, plugin := range apks {
			taskQueue <- CopyTask{
				//ChannelId: channel,
				SrcPath: plugin,
				DstPath: channel,
			}
		}
	}
	close(taskQueue)
	wg.Wait()
	return nil
}
func CopyApk(gameDirPath string, apks []string, targetDirs map[string]string) error {
	var wg sync.WaitGroup
	tm0 := time.Now().Unix()
	for _, dest := range targetDirs {
		wg.Add(1)
		go func(dst string) {
			defer wg.Done()
			utils.Copy(gameDirPath, filepath.Join(dst, "gameDir"), true)
		}(dest)
	}
	go func() {
		wg.Wait()
	}()
	errChan := make(chan error, len(apks)*len(targetDirs))
	println("母包目录拷贝完成")

	sem := make(chan struct{}, 32)
	var wg2 sync.WaitGroup
	for _, apk := range apks {
		for _, dir := range targetDirs {
			wg2.Add(1)
			sem <- struct{}{}
			go func(src, destDir string) {
				defer wg2.Done()
				defer func() { <-sem }()

				srcPath := filepath.Join("C:\\apktool", "sdk", "expand", src+"Dir")
				destPath := filepath.Join(destDir, src+"Dir")
				//if err := fastCopy(srcPath, destPath); err != nil {
				if err := utils.Copy(srcPath, destPath, false); err != nil {
					errChan <- fmt.Errorf("%s -> %s 失败: %v", srcPath, destPath, err)
					return
				}
				errChan <- nil
			}(apk, dir)
		}
	}
	go func() {
		wg2.Wait()
		close(errChan)
	}()
	// 检查错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	tm1 := time.Now().Unix()
	println("拷贝完成，耗时", tm1-tm0, "秒")
	return nil
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
		copyPlugin(task)
		log.Printf("拷贝完成, %s, %s, %v", task.SrcPath, task.DstPath, time.Since(start))
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
