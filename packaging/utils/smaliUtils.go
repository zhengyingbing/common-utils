package utils

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/packaging/models"
	"strconv"
	"strings"
	"sync"
)

func SmaliMap(path string, logger models.LogCallback) map[string]string {
	smaliPathMap := make(map[string]string)
	path = filepath.Join(path, "smali")
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			logger.LogDebug("遍历smali时异常：", err.Error())
			panic(err.(interface{}))
		}
		if info.IsDir() {
			return err
		}
		name := info.Name()
		if strings.HasSuffix(name, ".smali") {
			smaliPathMap[path] = name
		}
		return err
	})
	if err != nil {
		logger.LogDebug("遍历smali结果异常：", err.Error())
		panic(err.(interface{}))
	}
	return smaliPathMap
}

func MergeSmali(pluginPath, gamePath, pluginName, rules string, entries []os.DirEntry, logger models.LogCallback) {

	if utils.Exist(filepath.Join(pluginPath, "META-INF")) {
		utils.Remove(filepath.Join(pluginPath, "META-INF"))
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), "smali") {
			priority := strings.Contains(rules, entry.Name())
			//smali合并时全部合并到母包的“主smali”中
			_ = utils.Move(filepath.Join(pluginPath, entry.Name()), filepath.Join(gamePath, "smali"), priority)
		}
	}
	logger.LogVerbose(pluginName, "smali合并完成")
}

/**
 * gamePackagePath:包名路径 com\hoolai\xx\xx\xx
 * filePath:R文件路径 C:\apktool\test\smali\com\a\b\R$attr.smali
 * name:R文件名称 R$attr.smali
 */
func ReplaceRFile(gamePackagePath, filePath, name string, logger models.LogCallback) {
	name = strings.Split(name, ".")[0]
	//只取 attr
	name = strings.Split(name, "$")[1] //attr
	//logger.LogDebug("文件名称：", name)

	p := strings.Split(filePath, "smali")
	pluginPackagePath := p[1][1 : len(p[1])-1]
	pluginPackagePath = filepath.ToSlash(pluginPackagePath)

	utils.CreateDir(pluginPackagePath)
	//logger.LogDebug("R文件包名路径：", pluginPackagePath)
	var dstFilePath = filepath.ToSlash(gamePackagePath) + "/R_hoolai$" + name
	//logger.LogDebug("修改后的R文件路径：", dstFilePath)
	var tS = ".class public final L" + pluginPackagePath + ";\n" +
		".super L" + dstFilePath + ";\n" +
		".source \"R.java\"\n" +
		"\n\n" +
		"# annotations\n" +
		".annotation system Ldalvik/annotation/EnclosingClass;\n" +
		"	value = L" + filepath.ToSlash(filepath.Dir(pluginPackagePath)) + "/R;\n" +
		".end annotation\n" +
		"\n" +
		".annotation system Ldalvik/annotation/InnerClass;\n" +
		"	accessFlags = 0x19\n" +
		"	name = \"" + name + "\"\n" +
		".end annotation\n\n" +
		"# direct methods\n" +
		".method public constructor <init>()V\n" +
		"    .locals 0\n" +
		"\n" +
		"    .line 6\n" +
		"    invoke-direct {p0}, L" + dstFilePath + ";-><init>()V\n" +
		"\n" +
		"    return-void\n" +
		".end method\n"
	mgr := utils.CreateFileMgr()
	mgr.RemoveFile(filePath)
	mgr.WriteFile(filePath, tS)
	mgr.End(filePath)
}

func SubSmali(gamePath, channelId string, logger models.LogCallback) {
	defaultCounters := 50000
	val := models.GetServerDynamic(channelId)[models.DexMethodCounters]
	if val != "" {
		counter, _ := strconv.Atoi(val)
		defaultCounters = counter
	}
	logger.LogDebug(defaultCounters)
}

func BuildDex(java, smaliJar, gamePath, channelId string, logger models.LogCallback) {
	// 解析 AndroidManifest.xml
	classes, uniqueMap := CoreComponents(gamePath)
	//获取主dex文件名单（四大组件的类引用）
	mainMap := CoreDependencies(gamePath, classes, uniqueMap, logger)
	//统计四大组件及其类引用之外的smali文件
	smaliMap := getSmaliFile(gamePath, mainMap, logger)
	//smali方法数计算，拆分smali文件
	dexMap := smaliSubPackage(gamePath, channelId, logger, smaliMap)
	//proc := NewProcessor()
	//dexMap := proc.GetDexInfo(gamePath, channelId, logger, smaliMap)
	for smaliPath, dexPath := range dexMap {

		var buildDex = java + " -Dfile.encoding=utf-8 -jar " + smaliJar + " a " + smaliPath + " -o " + dexPath
		logger.LogDebug("执行命令:" + buildDex)
		err := utils.ExecuteShell(buildDex)
		if err != nil {
			logger.LogDebug("Error:" + err.Error())
			panic(err.(interface{}))
		} else {
			logger.LogDebug(dexPath, "构建成功")
		}

	}
}

func CoreDependencies(gamePath string, classes []string, uniqueMap map[string]bool, logger models.LogCallback) map[string]bool {
	for _, filePath := range classes {
		var url = filepath.Join(gamePath, "smali", filePath+".smali")
		file, err := os.Open(url)
		if err != nil {
			logger.LogDebug("无法打开文件：" + err.Error())
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, ".implements") {
				start := strings.Index(line, "L")
				end := strings.Index(line, ";")
				if start != -1 && end != -1 && start < end {
					appUrl := strings.Replace(line[start+1:end], "/", utils.Symbol(), -1)
					// 提取类名
					classes = append(classes, appUrl)
					uniqueMap[appUrl] = true
				}
			}
		}
	}
	return uniqueMap
}

func getSmaliFile(root string, mainMap map[string]bool, logger models.LogCallback) map[string]string {
	smaliFilePathMaps := make(map[string]string)
	err := filepath.Walk(filepath.Join(root, "smali"), func(path string, info os.FileInfo, err error) error {

		if err != nil {
			logger.LogDebug("Error:" + err.Error())
			panic(err.(interface{}))
		}

		if info.IsDir() {
			return err
		}
		var end = 0
		name := info.Name()

		start := strings.Index(path, "smali") + 6
		if strings.Contains(path, "$") {
			end = strings.Index(path, "$")
		} else {
			end = strings.LastIndex(path, ".smali")
		}
		if start != -1 && end != -1 && start < end {
			if !mainMap[path[start:end]] {
				smaliFilePathMaps[path] = name
			}
		}
		return err
	})

	if err != nil {
		logger.LogDebug("Error:" + err.Error())
		panic(err.(interface{}))
	}

	return smaliFilePathMaps
}

// 开始分包处理
func smaliSubPackage(buildApkDir, channelId string, logger models.LogCallback, smaliFilePathMaps map[string]string) map[string]string {
	smaliMap := map[string]string{}
	//1.计算smali文件方法数并重写到本地
	logger.LogDebug("start statistics smali All methods and write localSmaliFile.txt")

	//默认阈值
	dexMethodCounters := 65000
	val := models.GetServerDynamic(channelId)["dexMethodCounters"]
	if val != "" {
		mtc, _ := strconv.Atoi(val)
		dexMethodCounters = mtc
	}
	logger.LogDebug("设置的方法数阈值：", dexMethodCounters)
	//累计引用方法数
	currentInvokeMethodCount := 0
	//当前smali文件索引
	currentFolderIndex := 2
	dexMap := map[string]string{}
	for smaliFilePath, fileName := range smaliFilePathMaps {

		//获取单个文件的方法数
		invokeCount, _ := countMethodsInSmali(fileName, smaliFilePath, logger)

		if (currentInvokeMethodCount + invokeCount) > dexMethodCounters {
			logger.LogDebug(fmt.Sprintf("%s%d", "smali_classes", currentFolderIndex)+"方法数 = ", currentInvokeMethodCount)
			smaliPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d", "smali_classes", currentFolderIndex))
			dexPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d%s", "classes", currentFolderIndex, ".dex"))
			//中间的dex
			dexMap[smaliPath] = dexPath
			currentFolderIndex++
			currentInvokeMethodCount = 0
		}

		//移动文件到目标文件夹
		destNewPath := strings.Replace(smaliFilePath, filepath.Join(buildApkDir, "smali"), filepath.Join(buildApkDir, fmt.Sprintf("%s%d", "smali_classes", currentFolderIndex)), -1)
		smaliMap[smaliFilePath] = destNewPath

		currentInvokeMethodCount += invokeCount
	}

	//开始拷贝
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 10)
	// 遍历 smaliMap
	for srcFile, destNewPath := range smaliMap {
		wg.Add(1)
		sem <- struct{}{}
		go func(path string, targetPath string) {
			defer wg.Done()
			if err := utils.Move(path, targetPath, true); err != nil {
				mu.Lock()
				err = fmt.Errorf("Move failed for %s to %s: %v", path, targetPath, err)
				mu.Unlock()
			}
			<-sem
		}(srcFile, destNewPath)
	}
	// 等待所有 goroutine 完成
	wg.Wait()

	smaliPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d", "smali_classes", currentFolderIndex))
	dexPath := filepath.Join(buildApkDir, fmt.Sprintf("%s%d%s", "classes", currentFolderIndex, ".dex"))
	//最后一个dex
	dexMap[smaliPath] = dexPath

	//主dex
	dexMap[filepath.Join(buildApkDir, "smali")] = filepath.Join(buildApkDir, "classes.dex")

	logger.LogDebug("开始65535分包处理完成")
	return dexMap
}

// 获取单个文件的定义方法数及引用方法数
// R、内部类、BuildConfig不参与计数
func countMethodsInSmali(fileName string, smaliFile string, logger models.LogCallback) (int, error) {
	//|| strings.EqualFold(fileName, "BuildConfig") || strings.EqualFold(fileName, "R")
	//匿名内部类不参与计算
	if regexAnonymousClass.MatchString(fileName) || strings.EqualFold(fileName, "R") {
		return 0, fmt.Errorf("匿名内部类及R类不参与计算")
	}
	uniqueMap := make(map[string]bool)

	// 打开文件
	file, err := os.Open(smaliFile)
	if err != nil {
		return 0, fmt.Errorf("无法打开文件: %v", err)
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
		return 0, fmt.Errorf("读取文件时出错: %v", err)
	}
	return len(uniqueMap), nil
}
