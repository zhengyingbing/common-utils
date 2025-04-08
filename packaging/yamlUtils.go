package packaging

import (
	"gopkg.in/yaml.v3"
	"os"
)

/**
 * src:插件yaml
 * dst:母包yaml
 */
func MergeYaml(src, dst string) error {
	println("开始执行MergeYaml----")
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
