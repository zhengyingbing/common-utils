package utils

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

/**
 * src:插件yaml
 * dst:母包yaml
 */
func MergeYaml(src, dst string) error {

	srcResult, err := readYaml[map[string]interface{}](src)
	if err != nil {
		return err
	}

	dstResult, err := readYaml[map[string]interface{}](dst)
	if err != nil {
		return err
	}

	if _, ok := srcResult["doNotCompress"]; ok {
		srcCompress := srcResult["doNotCompress"].([]interface{})
		dstCompress := dstResult["doNotCompress"].([]interface{})

		collect := append(srcCompress, dstCompress...)
		dstResult["doNotCompress"] = Duplicate(collect)
	}

	if _, ok := srcResult["unknownFiles"]; ok {
		srcUnknownFiles := srcResult["unknownFiles"].(map[string]interface{})
		dstUnknownFiles := dstResult["unknownFiles"].(map[string]interface{})
		for key, val := range srcUnknownFiles {
			dstUnknownFiles[key] = val
		}

		dstResult["unknownFiles"] = dstUnknownFiles
	}
	marshal, err := yaml.Marshal(dstResult)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, marshal, os.ModePerm)
}

// 替换版本号
func ReplaceStringYaml(src, minSdkVersion, targetSdkVersion string) error {

	file, err := os.Open(src)
	if err != nil {
		fmt.Printf("无法打开文件: %v\n", err)
		return err
	}
	scanner := bufio.NewScanner(file)
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "minSdkVersion") {
			v := strings.Split(line, ":")[1]
			newLine := strings.Replace(line, v, " '"+minSdkVersion+"'", -1)
			lines = append(lines, newLine)
		} else if strings.Contains(line, "targetSdkVersion") {
			v := strings.Split(line, ":")[1]
			newLine := strings.Replace(line, v, " '"+targetSdkVersion+"'", -1)
			lines = append(lines, newLine)
		} else {
			lines = append(lines, line)
		}
	}

	output, _ := os.Create(src)
	defer output.Close()
	writer := bufio.NewWriter(output)
	for _, line := range lines {
		_, err = writer.WriteString(line + "\n")
		if err != nil {
			fmt.Printf("写入文件错误: %v\n", err)
			return err
		}
	}

	writer.Flush()
	return nil
}

func readYaml[T any](yamlPath string) (T, error) {
	var data T
	buf, err := os.ReadFile(yamlPath)
	if err != nil {
		return data, err
	}
	err = yaml.Unmarshal(buf, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

/**
 * 泛型哈希法去重元素
 * 类型安全，时间复杂度O(n)
 */
func Duplicate[T comparable](data []T) []T {
	if len(data) == 0 {
		return nil
	}
	dataMap := make(map[T]struct{}, len(data))
	result := make([]T, 0, len(data))

	for _, v := range data {
		if _, exist := dataMap[v]; !exist {
			dataMap[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
