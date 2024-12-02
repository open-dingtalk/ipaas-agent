package plugins

import (
	"context"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

type VersionPlugin struct {
	Name            string `json:"name"`
	ProtocolVersion string `json:"protocol_version"`
}

func NewVersionPlugin() *VersionPlugin {
	return &VersionPlugin{
		Name:            "version_plugin",
		ProtocolVersion: "2.0",
	}
}

func (p *VersionPlugin) Init() error {
	// 初始化插件，例如读取配置
	return nil
}

func (p *VersionPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	// 处理消息，例如返回版本信息
	return v1.NewSuccessDataFrameResponse(&p), nil
}

func (p *VersionPlugin) Close() error {
	// 关闭插件
	logger.Log1.WithField("plugin", p.Name).Info("插件已关闭")
	return nil
}
