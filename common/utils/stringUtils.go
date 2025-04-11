package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func Map2Struct(data map[string]interface{}, obj interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &obj)
}

// 将任何数据类型转成JSON切片
func Struct2Json(data interface{}) string {
	marshal, _ := json.Marshal(data)
	return string(marshal)
}

/**
 * escape: 是否对html中的特殊字符(<,>,&,',")进行转义为Unicode序列
 */
func Struct2EscapeJson(data interface{}, escape bool) string {
	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(escape)
	_ = encoder.Encode(data)
	return bf.String()
}

func Json2Struct(jsonStr string, obj interface{}) error {
	return json.Unmarshal([]byte(jsonStr), &obj)
}

func Hex2Dec(val string) int {
	val = strings.TrimSpace(val)
	val = strings.ReplaceAll(val, "0x", "")
	n, err := strconv.ParseUint(val, 16, 32)
	if err != nil {
		fmt.Println(err)
	}
	return int(n)
}

func Dec2Hex(val int) string {
	i := int64(val)
	s := strconv.FormatInt(i, 16)
	return "0x" + s
}
