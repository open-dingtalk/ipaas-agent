package client

import (
	"context"
	"sync"
	"time"

	"github.com/open-dingtalk/ipaas-agent/pkg/config"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/config/v1"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	"github.com/open-dingtalk/ipaas-agent/pkg/plugins"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	sdkLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
)

const (
	// 钉钉回调 topic
	callbackTopic = "/v1.0/ipaas/proxy/callback"
	// WebSocket 层空闲心跳间隔。SDK 默认 120s，穿过 NAT/防火墙时容易被静默掐断，
	// 这里压到 30s。
	keepAliveIdle = 30 * time.Second
	// 启动期建连失败后的重试间隔
	initialRetryInterval = 2 * time.Second
	maxRetryInterval     = time.Minute
)

type Client struct {
	conf          *v1.AuthClientConfig
	pluginManager *plugins.PluginManager
	streamClient  *client.StreamClient
	mu            sync.Mutex
}

func NewClient(conf *v1.AuthClientConfig, pm *plugins.PluginManager) *Client {
	return &Client{
		conf:          conf,
		pluginManager: pm,
	}
}

// Connect 建立与钉钉开放平台的连接。
//
// 首次建连失败时不会放弃：SDK 的自动重连只在首次成功之后才生效，而网络抖动、
// 代理未就绪、凭证下发延迟都会让启动期的第一次建连失败。此时把错误返回给调用方
// 用于提示，同时在后台按退避持续重试，避免整个网关因此退出。
func (c *Client) Connect(ctx context.Context) error {
	// 注册日志Logger
	sdkLogger.SetLogger(logger.Log2)

	err := c.start(ctx)
	if err == nil {
		return nil
	}
	go c.retryUntilConnected(ctx)
	return err
}

// start 按当前配置新建一个 StreamClient 并建连，成功后替换掉旧实例。
func (c *Client) start(ctx context.Context) error {
	auth := config.GetAuthClientConfig()
	options := []client.ClientOption{
		client.WithAppCredential(client.NewAppCredentialConfig(auth.ClientID, auth.ClientSecret)),
		client.WithOpenApiHost(auth.OpenAPIHost),
		client.WithExtras(map[string]string{}),
		client.WithKeepAlive(keepAliveIdle),
	}
	if auth.Proxy != "" {
		options = append(options, client.WithProxy(auth.Proxy))
		logger.Log1.WithField("proxy", auth.Proxy).Info("使用代理连接钉钉开放平台")
	}

	streamClient := client.NewStreamClient(options...)
	// 注册事件类型的处理函数 callback 是 goroutin
	streamClient.RegisterCallbackRouter(callbackTopic, c.handleServerMessage)
	if err := streamClient.Start(ctx); err != nil {
		logger.Log1.Errorf("连接到服务器失败: %v", err)
		return err
	}
	// 建连成功时进程可能已经在退出了，这条连接要立刻关掉，不能留给后台
	if err := ctx.Err(); err != nil {
		logger.Log1.Info("建连成功但进程正在退出，立即关闭连接")
		streamClient.Close()
		return err
	}

	c.mu.Lock()
	c.conf = auth
	c.streamClient = streamClient
	c.mu.Unlock()
	logger.Log1.Info("成功连接到服务器")
	return nil
}

// retryUntilConnected 指数退避地重试建连，直到成功或 ctx 结束。
func (c *Client) retryUntilConnected(ctx context.Context) {
	interval := initialRetryInterval
	for {
		select {
		case <-ctx.Done():
			logger.Log1.Info("退出连接重试")
			return
		case <-time.After(interval):
		}

		if c.connected() {
			return
		}
		if err := c.start(ctx); err == nil {
			return
		}

		interval *= 2
		if interval > maxRetryInterval {
			interval = maxRetryInterval
		}
		logger.Log1.WithField("retry_in", interval.String()).Warn("与服务器的连接仍未建立，稍后重试")
	}
}

func (c *Client) connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.streamClient != nil
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
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.streamClient != nil {
		c.streamClient.Close()
		c.streamClient = nil
	}
	logger.Log1.Info("已断开与服务器的连接")
	c.pluginManager.CloseAll()
}
