package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

type HTTPPlugin struct {
	// 可以有插件自己的配置
	pm *PluginManager
}

func NewHTTPPlugin() *HTTPPlugin {
	return &HTTPPlugin{}
}

func (p *HTTPPlugin) Init() error {
	// 初始化插件，例如读取配置
	logger.Log1.Info("HTTP插件已初始化")
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

// 针对不同版本的消息，可以有不同的处理逻辑
func (p *HTTPPlugin) handleV1(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	// 处理 v1 版本的逻辑
	logger.Log1.Info("处理 v1 版本的消息")
	var dataModel v1.IPaaSAgentProtocol
	err := json.Unmarshal([]byte(df.Data), &dataModel)
	if err != nil {
		return payload.NewErrorDataFrameResponse(err), err
	}

	// 二次路由插件
	for k, v := range dataModel.Body.HTTPRequest.Headers {
		if strings.ToUpper(k) == "X-IPAAS-PLUGINNAME" {
			pluginName := v
			plugin, exists := p.pm.plugins[pluginName]
			if !exists {
				break
			}
			return plugin.HandleMessage(ctx, df)
		}
	}

	resp, err := v1.HandleHTTPRequest(dataModel.Body.HTTPRequest)
	if err != nil {
		return payload.NewErrorDataFrameResponse(err), err
	}
	dfResp := payload.NewSuccessDataFrameResponse()
	dfResp.SetData(string(resp))
	return dfResp, nil
}

func (p *HTTPPlugin) handleV2(_ context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	// 处理 v2 版本的逻辑
	logger.Log1.Info("处理 v2 版本的消息 TODO ...")
	// 示例处理逻辑
	// 你可以在这里添加具体的处理逻辑
	return &payload.DataFrameResponse{
		Data: "处理 v2 版本的响应数据",
	}, nil
}

func (p *HTTPPlugin) Close() error {
	// 关闭插件
	logger.Log1.Info("HTTP插件已关闭")
	return nil
}
