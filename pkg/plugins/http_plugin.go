package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

type HTTPPlugin struct {
	// 可以有插件自己的配置
	pm   *PluginManager
	Name string
}

func NewHTTPPlugin() *HTTPPlugin {
	return &HTTPPlugin{
		Name: "http_plugin",
	}
}

func (p *HTTPPlugin) Init() error {
	// 初始化插件，例如读取配置
	logger.Log1.WithField("plugin", p.Name).Info("HTTP插件已初始化")
	return nil
}

func (p *HTTPPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	// 初始化 Data
	dataVersion := df.GetDataVersion()
	switch dataVersion {
	case "1.0":
		return p.handleV1(ctx, df)
	case "2.0":
		return p.handleV2(ctx, df)
	default:
		return nil, fmt.Errorf("不支持的 dataVersion: %s", dataVersion)
	}
}

func getResponseBodyStr(response interface{}) (string, error) {
	switch v := response.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	default:
		// 尝试 JSON 序列化
		data, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("序列化响应失败: %v", err)
		}
		return string(data), nil
	}
}

// v1 协议兼容代码，兼容 HTTP 代理和插件
func (p *HTTPPlugin) handleV1(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	dataModel := df.GetDataModelV1()

	// 二次路由插件
	for k, v := range dataModel.Body.HTTPRequest.Headers {
		if strings.ToLower(k) == "x-ipaas-plugin-name" {
			pluginName := v
			plugin, exists := p.pm.plugins[pluginName]
			if !exists {
				break
			}

			r, err := plugin.HandleMessage(ctx, df)
			if err != nil {
				return payload.NewErrorDataFrameResponse(err), err
			}

			// 转换为 http 的响应
			resp, err := v1.GetResponseFromDataFrameResponse(r)
			if err != nil {
				return payload.NewErrorDataFrameResponse(err), err
			}

			bodyStr, err := getResponseBodyStr(resp)
			if err != nil {
				return payload.NewErrorDataFrameResponse(err), err
			}

			// resp = &http.Response{
			// 	StatusCode: r.Code,
			// 	Status:     http.StatusText(r.Code),
			// 	Proto:      "HTTP/1.1",
			// 	ProtoMajor: 1,
			// 	ProtoMinor: 1,
			// 	Header:     make(http.Header),
			// 	Body:       io.NopCloser(bytes.NewBufferString(bodyStr)),
			// }
			resp = &v1.HTTPResponse{
				StatusCode: r.Code,
				Status:     http.StatusText(r.Code),
				Proto:      "HTTP/1.1",
				Body:       bodyStr,
			}
			return v1.NewSuccessDataFrameResponse(resp), nil
		}
	}

	// 正常 http 请求
	resp, err := v1.HandleHTTPRequest(dataModel.Body.HTTPRequest)
	callbackResponse := &CallbackResponse{
		Response: resp,
	}
	if err != nil {
		return payload.NewErrorDataFrameResponse(err), err
	}
	dfResp := payload.NewSuccessDataFrameResponse()
	dfResp.SetJson(callbackResponse)
	return dfResp, nil
}

// v2 仅仅处理HTTP请求，插件在外面已经被路由
func (p *HTTPPlugin) handleV2(_ context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	// 处理 v2 版本的逻辑
	logger.Log1.Info("处理 v2 版本的消息")
	// 示例处理逻辑
	// 你可以在这里添加具体的处理逻辑
	return &payload.DataFrameResponse{
		Data: "处理 v2 版本的响应数据",
	}, nil
}

func (p *HTTPPlugin) Close() error {
	// 关闭插件
	logger.Log1.WithField("plugin", p.Name).Info("插件已关闭")
	return nil
}
