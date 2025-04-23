package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	utils2 "sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strconv"
	"strings"
	"sync"
)

// 匹配内部类
var regexAnonymousClass = regexp.MustCompile(`.*\$\d+`)

// 匹配方法调用指令的正则表达式
var regexInvokeMethod = regexp.MustCompile(`invoke-(.*)L[^(java)(android)].*`)

type ClassInfo struct {
	smaliFilePath string
	invokeCount   int
}

type Processor struct {
	dexMap       map[string]string
	smaliMap     map[string]string
	classInfoMap map[string]*ClassInfo
	mu           sync.RWMutex
}

func NewProcessor() *Processor {
	return &Processor{
		dexMap:       make(map[string]string),
		smaliMap:     make(map[string]string),
		classInfoMap: make(map[string]*ClassInfo),
		mu:           sync.RWMutex{},
	}
}

func (p *Processor) GetDexInfo(buildApkDir, channelId string, logger models.LogCallback, smaliFilePathMaps map[string]string) map[string]string {
	//1.计算smali文件方法数并重写到本地
	logger.LogDebug("start statistics smali All methods and write localSmaliFile.txt")

	//默认阈值
	dexMethodCounters := 65000
	val := models.GetServerDynamic(channelId)[models.DexMethodCounters]
	if val != "" {
		mtc, _ := strconv.Atoi(val)
		dexMethodCounters = mtc
	}
	logger.LogDebug("设置的方法数阈值：", dexMethodCounters)
	//累计引用方法数
	currentInvokeMethodCount := 0
	//当前smali文件索引
	currentFolderIndex := 2

	var wg0 sync.WaitGroup

	for smaliFilePath, fileName := range smaliFilePathMaps {
		wg0.Add(1)
		go func(f string, fn string, log models.LogCallback) {
			defer wg0.Done()
			//获取单个文件的方法数
			p.CountMethodsInSmali(f, fn, log)
		}(fileName, smaliFilePath, logger)
	}
	wg0.Wait()
	logger.LogDebug("统计smali文件方法数完成")
	if p.classInfoMap != nil && len(p.classInfoMap) > 0 {
		for _, classInfo := range p.classInfoMap {
			invokeCount := classInfo.invokeCount
			if (currentInvokeMethodCount + invokeCount) > dexMethodCounters {
				logger.LogDebug("方法数 = ", currentInvokeMethodCount)
				smaliPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d", "smali_classes", currentFolderIndex))
				dexPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d%s", "classes", currentFolderIndex, ".dex"))
				//中间的dex
				p.dexMap[smaliPath] = dexPath
				currentFolderIndex++
				currentInvokeMethodCount = 0
			}

			//移动文件到目标文件夹
			destNewPath := strings.Replace(classInfo.smaliFilePath, filepath.Join(buildApkDir, "smali"),
				filepath.Join(buildApkDir, fmt.Sprintf("%s%d", "smali_classes", currentFolderIndex)), -1)
			p.smaliMap[classInfo.smaliFilePath] = destNewPath
			//err2 := util.Move(smaliFilePath, destNewPath)
			//if err2 != nil {
			//	logger.LogDebug(err2.Error())
			//}
			currentInvokeMethodCount += invokeCount
		}
	}

	logger.LogDebug("开始拷贝文件操作")

	// 遍历 smaliMap
	for srcFile, destNewPath := range p.smaliMap {
		err := utils2.Move(srcFile, destNewPath, true)
		if err != nil {
			logger.LogDebug(err.Error())
			logger.LogDebug("移动路径", srcFile, destNewPath)
			panic(err)
		}
	}
	// 等待所有 goroutine 完成
	// wg.Wait()

	smaliPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d", "smali_classes", currentFolderIndex))
	dexPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d%s", "classes", currentFolderIndex, ".dex"))
	//最后一个dex
	p.dexMap[smaliPath] = dexPath

	//主dex
	p.dexMap[filepath.Join(buildApkDir, "smali")] = filepath.Join(buildApkDir, "classes.dex")

	logger.LogDebug("开始65535分包处理完成")
	return p.dexMap
}

// 获取单个文件的定义方法数及引用方法数
// R、内部类、BuildConfig不参与计数
func (p *Processor) CountMethodsInSmali(fileName string, smaliFile string, logger models.LogCallback) error {
	//|| strings.EqualFold(fileName, "BuildConfig") || strings.EqualFold(fileName, "R")
	//匿名内部类不参与计算
	if regexAnonymousClass.MatchString(fileName) || strings.EqualFold(fileName, "R") {
		return fmt.Errorf("匿名内部类及R类不参与计算")
	}
	uniqueMap := make(map[string]bool)

	// 打开文件
	file, err := os.Open(smaliFile)
	if err != nil {
		return fmt.Errorf("无法打开文件: %v", err)
	}
	defer file.Close() // 确保文件关闭
	// 使用 bufio.Scanner 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		//匹配引用方法
		matches := regexInvokeMethod.FindStringSubmatch(line)

		if strings.HasPrefix(line, ".method") {
			//methodCount++
			uniqueMap[line] = true
		}
		if len(matches) > 1 {
			uniqueMap[line] = true
		}
	}
	// 检查扫描过程中是否出错
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件时出错: %v", err)
	}

	info := &ClassInfo{}
	info.smaliFilePath = smaliFile
	info.invokeCount = len(uniqueMap)
	p.mu.Lock()
	p.classInfoMap[smaliFile] = info
	p.mu.Unlock()
	return nil
}
