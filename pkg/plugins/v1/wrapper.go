package v1

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
)

type DFWrap struct {
	*payload.DataFrame
	dataJson    map[string]interface{}
	DataModelv1 *IPaaSAgentProtocol
}

type CallbackResponse struct {
	Response interface{} `json:"response"`
}

func (df *DFWrap) GetDataModelV1() *IPaaSAgentProtocol {
	if df.DataModelv1 == nil {
		err := json.Unmarshal([]byte(df.Data), &df.DataModelv1)
		if err != nil {
			logger.Log1.Errorf("解析 DataFrame.Data 出错: %v", err)
			return nil
		}
	}
	return df.DataModelv1
}

func (df *DFWrap) GetDataJson() map[string]interface{} {
	if df.dataJson == nil {
		err := json.Unmarshal([]byte(df.DataFrame.Data), &df.dataJson)
		if err != nil {
			logger.Log1.Errorf("解析 DataFrame.Data 出错: %v", err)
			return nil
		}
	}
	return df.dataJson
}

// 获取 IPaaS DataFrame 的版本信息 默认 1.0
func (df *DFWrap) GetDataVersion() string {
	dataJson := df.GetDataJson()
	// 2.0 版本 specVersion 字段在 data 里
	if specVersion, exists := dataJson["specVersion"].(string); exists {
		return specVersion
	}
	// 1.0 版本 specVersion 字段在 headers 里
	if headers, ok := dataJson["headers"].(map[string]interface{}); ok {
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

// v1 协议: IPaaSAgentProtocol pkg/plugins/v1/proxy_mysql.go
func (df *DFWrap) getPluginNameV1() string {
	dataJson := df.GetDataJson()
	// 1.0 从 header 里获取
	if headers, ok := dataJson["headers"].(map[string]interface{}); ok {
		// 这层 headers 是连接平台包装的
		if protocolType, ok := headers["type"].(string); ok {
			switch strings.ToLower(protocolType) {
			case "http":
				return "http_plugin"
			case "mysql":
				return "proxy_mysql_plugin"
			}
		}
	}
	// 如果都没有，则返回 http_plugin
	return "http_plugin"
}

func (df *DFWrap) getPluginNameV2() string {
	dataJson := df.GetDataJson()
	if pluginName, exists := dataJson["pluginName"].(string); exists {
		return pluginName
	}
	return "http_plugin"
}

// 使用示例:
//
//	type MyData struct {
//	    Field1 string `json:"field1"`
//	    Field2 int    `json:"field2"`
//	}
//
// data, err := df.GetPluginDataWithType(reflect.TypeOf(MyData{}))
//
//	if err != nil {
//	    // 处理错误
//	}
//
// myData := data.(*MyData)
func (df *DFWrap) GetPluginDataWithType(t reflect.Type) (interface{}, error) {
	// 获取原始数据
	data := df.GetPluginData()
	if data == nil {
		return nil, fmt.Errorf("no data found for version %s", df.GetDataVersion())
	}

	// 创建目标类型的新实例
	targetValue := reflect.New(t).Interface()

	// 检查 data 类型是否为 string
	if strData, ok := data.(string); ok {
		// 如果是字符串，直接使用
		if err := json.Unmarshal([]byte(strData), targetValue); err != nil {
			return nil, fmt.Errorf("unmarshal string error: %w", err)
		}
		return targetValue, nil
	}

	// 如果不是字符串，先序列化为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	// 将 JSON 反序列化为目标类型
	if err := json.Unmarshal(jsonData, targetValue); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return targetValue, nil
}

// 获取插件数据
func (df *DFWrap) GetPluginData() interface{} {
	dataVersion := df.GetDataVersion()
	switch dataVersion {
	case "1.0":
		return df.getPluginDataV1()
	case "2.0":
		return df.getPluginDataV2()
	default:
		return nil
	}
}

func (df *DFWrap) getPluginDataV1() interface{} {
	// 1.0 从 http body 里获取
	return df.GetDataModelV1().Body.HTTPRequest.Body
}

func (df *DFWrap) getPluginDataV2() interface{} {
	// 2.0 直接从 data 里获取
	if data, exists := df.GetDataJson()["data"]; exists {
		return data
	}
	return nil
}

// 创建一个由 CallbackResponse 包装的 DataFrameResponse
func NewSuccessDataFrameResponse(data interface{}) *payload.DataFrameResponse {
	cr := CallbackResponse{
		Response: data,
	}
	resp := payload.NewSuccessDataFrameResponse()
	resp.SetJson(cr)
	return resp
}

func GetResponseFromDataFrameResponse(df *payload.DataFrameResponse) (interface{}, error) {
	var cr CallbackResponse
	if err := json.Unmarshal([]byte(df.Data), &cr); err != nil {
		return nil, err
	}
	return cr.Response, nil
}
