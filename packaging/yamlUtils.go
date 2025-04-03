package packaging

import (
	"gopkg.in/yaml.v3"
	"os"
)

func MergeYaml(src, dst string) error {
	srcResult, err := readYaml[map[string]any](src)
	if err != nil {
		return err
	}

	dstResult, err := readYaml[map[string]any](dst)
	if err != nil {
		return err
	}

	if _, ok := srcResult["doNotCompress"]; ok {
		srcCompress := srcResult["doNotCompress"].([]interface{})
		dstCompress := dstResult["doNotCompress"].([]interface{})
		collect := append(srcCompress, dstCompress)
		srcResult["doNotCompress"] = Duplicate(collect)
	}

	if _, ok := srcResult["unknownFiles"]; ok {
		srcCompress := srcResult["unknownFiles"].([]interface{})
		dstCompress := dstResult["unknownFiles"].([]interface{})
		collect := append(srcCompress, dstCompress)
		srcResult["unknownFiles"] = Duplicate(collect)
	}
	marshal, err := yaml.Marshal(srcResult)
	if err != nil {
		return err
	}
	return os.WriteFile(src, marshal, os.ModePerm)
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
	dataMap := make(map[T]struct{})
	result := make([]T, 0, len(data))

	for _, v := range data {
		if _, exist := dataMap[v]; !exist {
			dataMap[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}
