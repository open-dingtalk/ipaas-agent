package client

import (
	"context"

	"github.com/open-dingtalk/ipaas-agent/pkg/config"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/config/v1"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	"github.com/open-dingtalk/ipaas-agent/pkg/plugins"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	sdkLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
)

type Client struct {
	conf          *v1.AuthClientConfig
	pluginManager *plugins.PluginManager
	streamClient  *client.StreamClient
}

func NewClient(conf *v1.AuthClientConfig, pm *plugins.PluginManager) *Client {
	return &Client{
		conf:          conf,
		pluginManager: pm,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	// 初始化与服务器的连接
	// 注册日志Logger
	sdkLogger.SetLogger(logger.Log2)
	auth := config.GetAuthClientConfig()
	c.streamClient = client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(auth.ClientID, auth.ClientSecret)),
		client.WithOpenApiHost(auth.OpenAPIHost),
		client.WithExtras(map[string]string{}),
	)

	// 注册事件类型的处理函数 callback 是 goroutin
	c.streamClient.RegisterCallbackRouter("/v1.0/ipaas/proxy/callback", c.handleServerMessage)
	err := c.streamClient.Start(ctx)
	if err != nil {
		logger.Log1.Errorf("连接到服务器失败: %v", err)
		return err
	}
	logger.Log1.Info("成功连接到服务器")
	return nil
}

func (c *Client) handleServerMessage(ctx context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
	// 根据消息类型选择插件处理
	logger.Log1.Infof("收到服务器消息: %v", df)
	response, err := c.pluginManager.HandleMessage(ctx, df)
	if err != nil {
		logger.Log1.WithField("错误", err).Errorf("处理消息失败")
		return nil, err
	}
	return response, nil
}

func (c *Client) Disconnect() {
	if c.streamClient != nil {
		c.streamClient.Close()
	}
	logger.Log1.Info("已断开与服务器的连接")
	c.pluginManager.CloseAll()
}
