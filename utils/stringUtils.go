package utils

import "encoding/json"

func Map2Struct(data map[string]interface{}, obj interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &obj)
}
