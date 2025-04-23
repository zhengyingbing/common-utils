package models

import (
	"errors"
	"path/filepath"
	"sdk.wdyxgames.com/gitlab/platform-project/package/package-core/common/utils"
)

var dynamicCfgMap map[string]map[string]string

const (
	AppName           = "appName"
	IconName          = "iconName"
	BundleId          = "bundleId"
	Orientation       = "orientation"
	KeystoreAlias     = "keystoreAlias"
	KeystorePass      = "keystorePass"
	KeyPass           = "keyPass"
	SignVersion       = "signVersion"
	TargetSdkVersion  = "targetSdkVersion"
	DexMethodCounters = "dexMethodCounters"
)

func SetServerDynamic(channelId string, cfg map[string]string) {
	if dynamicCfgMap == nil {
		dynamicCfgMap = make(map[string]map[string]string, 0)
	}
	dynamicCfgMap[channelId] = cfg
}

func GetServerDynamic(channelId string) map[string]string {
	return dynamicCfgMap[channelId]
}

func OperateDynamic(gamePath, channelId string, configs []DynamicConfig, logger LogCallback) error {
	for _, item := range configs {
		filePath := filepath.Join(gamePath, item.ContentPath)
		for _, operate := range item.Operates {
			serverDynamic := GetServerDynamic(channelId)
			if serverDynamic == nil {
				logger.LogDebug("渠道", channelId, "未获取到管理台配置！")
				return errors.New("get server dynamic config failed！")
			}
			switch operate.Operate {
			case CopyClass:
			case Copy:
			case Replace:
				replace(operate.Config, filePath, serverDynamic, logger)
			case Delete:
			case Move:
				move(operate.Config, gamePath, logger)
			}
		}
	}
	return nil
}

/**
 * 将服务端的配置替换到dynamic_config中对应的标签中
 * dynamic_config的config对象
 * 母包路径
 * 服务端配置对象
 */
func replace(config map[string]interface{}, filePath string, dynamic map[string]string, logger LogCallback) {
	cfg := ReplaceConfig{}
	err := utils.Json2Struct(utils.Struct2Json(config), &cfg)
	if err != nil {
		logger.LogDebug("dynamic_config replace convert failed")
		return
	}
	tarVal := cfg.ConstantVal
	//如果server=false，则cfg.VarName替换成cfg.ConstantVal
	if cfg.Server {
		if v, ok := dynamic[cfg.VarName]; ok {
			tarVal = v
		}
	}
	err = utils.ReplaceFile(filePath, cfg.TarName, tarVal)
	if err != nil {
		logger.LogDebug("dynamic_config replace failed: " + err.Error())
	}
}

func move(config map[string]interface{}, gamePath string, logger LogCallback) {
	cfg := MoveConfig{}
	err := utils.Json2Struct(utils.Struct2Json(config), &cfg)
	if err != nil {
		logger.LogDebug("dynamic_config move convert failed")
		return
	}
	_ = utils.Move(filepath.Join(gamePath, cfg.Source), filepath.Join(gamePath, cfg.Target), true)
}
