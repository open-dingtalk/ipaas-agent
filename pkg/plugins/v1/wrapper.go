package v1

import (
	"encoding/json"
	"strings"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
)

type DFWrap struct {
	*payload.DataFrame
	DataJson map[string]interface{}
}

// 获取 IPaaS DataFrame 的版本信息 默认 1.0
func (df *DFWrap) GetDataVersion() string {
	if df.DataJson == nil {
		err := json.Unmarshal([]byte(df.DataFrame.Data), &df.DataJson)
		if err != nil {
			logger.Log1.Errorf("解析 DataFrame.Data 出错: %v", err)
			return "none"
		}
	}
	if specVersion, exists := df.DataJson["specVersion"].(string); exists {
		return specVersion
	}
	if headers, ok := df.DataJson["headers"].(map[string]interface{}); ok {
		if pluginName, exists := headers["specVersion"].(string); exists {
			return pluginName
		}
	}
	return "unknown version"
}

func (df *DFWrap) GetPluginName() string {
	dataVersion := df.GetDataVersion()
	switch dataVersion {
	case "1.0":
		return df.getPluginNameV1()
	case "2.0":
		return df.getPluginNameV2()
	default:
		return "http_plugin"
	}
}

func (df *DFWrap) getPluginNameV1() string {
	// 1.0 从 header 里获取
	if headers, ok := df.DataJson["headers"].(map[string]interface{}); ok {
		// 这层 headers 是连接平台包装的
		if protocolType, ok := headers["type"].(string); ok {
			switch strings.ToLower(protocolType) {
			case "http":
				return "http_plugin"
			case "mysql":
				return "mysql_plugin"
			}
		}
		// 从 headers 中获取 pluginName
		if pluginName, ok := headers["pluginName"].(string); ok {
			return pluginName
		}
	}
	// 如果都没有，则返回 http_plugin
	return "http_plugin"
}

func (df *DFWrap) getPluginNameV2() string {
	if pluginName, exists := df.DataJson["pluginName"].(string); exists {
		return pluginName
	}
	return "http_plugin"
}
