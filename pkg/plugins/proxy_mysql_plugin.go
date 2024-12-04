package plugins

import (
	"context"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

type ProxyMySQLPlugin struct {
	// 可以有插件自己的配置
	Name string
}

func NewProxyMySQLPlugin() *ProxyMySQLPlugin {
	return &ProxyMySQLPlugin{
		Name: "proxy_mysql_plugin",
	}
}

func (p *ProxyMySQLPlugin) Init() error {
	// 初始化插件，例如读取配置
	logger.Log1.
		WithField("插件名", p.Name).
		Info("插件已初始化")
	return nil
}

func (p *ProxyMySQLPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	res, err := v1.HandleMySQLProxyRequest(df.GetDataModelV1())
	if err != nil {
		return payload.NewErrorDataFrameResponse(err), err
	}
	return v1.NewSuccessDataFrameResponseV1(res), nil
}

func (p *ProxyMySQLPlugin) Close() error {
	// 关闭插件
	logger.Log1.WithField("plugin", p.Name).Info("插件已关闭")
	return nil
}
