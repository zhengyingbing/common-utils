package utils

import "encoding/json"

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

func Json2Struct(jsonStr string, obj interface{}) error {
	return json.Unmarshal([]byte(jsonStr), &obj)
}
