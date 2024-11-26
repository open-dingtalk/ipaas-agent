package plugins

import (
	"context"
	"fmt"
	"sync"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

type Plugin interface {
	Init() error
	// HandleMessage(ctx context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error)
	HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error)
	Close() error
}

type PluginManager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

func (pm *PluginManager) LoadPlugins() error {
	// 加载插件，可以使用反射或手动注册
	httpPlugin := NewHTTPPlugin()
	err := httpPlugin.Init()
	httpPlugin.pm = pm
	if err != nil {
		logger.Log1.Errorf("初始化 HTTP 插件失败: %v", err)
	}
	pm.RegisterPlugin("http_plugin", httpPlugin)

	versionPlugin := NewVersionPlugin()
	err = versionPlugin.Init()
	if err != nil {
		logger.Log1.Errorf("初始化 Version 插件失败: %v", err)
	}
	pm.RegisterPlugin("version_plugin", versionPlugin)

	return nil
}

func (pm *PluginManager) RegisterPlugin(name string, plugin Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.plugins[name] = plugin
	logger.Log1.Infof("插件已注册: %s", name)
}

func (pm *PluginManager) HandleMessage(ctx context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
	// 根据消息类型选择对应的插件处理
	dfWrap := &v1.DFWrap{
		DataFrame: df,
	}
	pluginName := dfWrap.GetPluginName()
	pm.mu.RLock()
	plugin, exists := pm.plugins[pluginName]
	pm.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("未找到对应的插件: %s", pluginName)
	}
	// event.NewSuccessResponse()
	return plugin.HandleMessage(ctx, dfWrap)
}
