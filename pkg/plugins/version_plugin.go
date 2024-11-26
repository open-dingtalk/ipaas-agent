package plugins

import (
	"context"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

type VersionPlugin struct {
}

func (p *VersionPlugin) GetVersion() string {
	return "v1.0.0"
}

func NewVersionPlugin() *VersionPlugin {
	return &VersionPlugin{}
}

func (p *VersionPlugin) Init() error {
	// 初始化插件，例如读取配置
	logger.Log1.Info("Version插件已初始化")
	return nil
}

func (p *VersionPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	// 处理消息，例如返回版本信息
	return &payload.DataFrameResponse{
		Code: 200,
		Headers: payload.DataFrameHeader{
			"ContentType": "application/json",
		},
		Message: "success",
		Data:    string(p.GetVersion()),
	}, nil
}

func (p *VersionPlugin) Close() error {
	// 关闭插件
	logger.Log1.Info("Version插件已关闭")
	return nil
}
