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

// 回调消息的响应
// github.com/open-dingtalk/dingtalk-stream-sdk-go@v0.9.0/plugin/plugin_handler.go
type CallbackResponse v1.CallbackResponse

type PluginManager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

func (pm *PluginManager) ReloadConfig() error {
	// Reinitialize all plugins
	pm.mu.RLock()
	for name, plugin := range pm.plugins {
		if err := plugin.Init(); err != nil {
			logger.Log1.Errorf("重新初始化插件 %s 失败: %v", name, err)
		}
	}
	pm.mu.RUnlock()

	return nil
}

func (pm *PluginManager) LoadPlugins() error {
	// 加载插件，可以使用反射或手动注册

	// 1. http 插件
	httpPlugin := NewHTTPPlugin()
	err := httpPlugin.Init()
	httpPlugin.pm = pm
	if err != nil {
		logger.Log1.Errorf("初始化 HTTP 插件失败: %v", err)
	}
	pm.RegisterPlugin(httpPlugin.Name, httpPlugin)

	// 2. 版本管理插件
	versionPlugin := NewVersionPlugin()
	err = versionPlugin.Init()
	if err != nil {
		logger.Log1.Errorf("初始化 Version 插件失败: %v", err)
	}
	pm.RegisterPlugin(versionPlugin.Name, versionPlugin)

	// 3. 旧 mysql 插件
	ProxyMySQLPlugin := NewProxyMySQLPlugin()
	err = ProxyMySQLPlugin.Init()
	if err != nil {
		logger.Log1.Errorf("初始化 ProxyMySQL 插件失败: %v", err)
	}
	pm.RegisterPlugin(ProxyMySQLPlugin.Name, ProxyMySQLPlugin)

	// 4. 新 mysql 插件
	mysqlPlugin := NewMySQLPlugin()
	err = mysqlPlugin.Init()
	if err != nil {
		logger.Log1.Errorf("初始化 MySQL 插件失败: %v", err)
	}
	pm.RegisterPlugin(mysqlPlugin.Name, mysqlPlugin)

	// 5. ms sql 插件
	mssqlPlugin := NewMSSQLPlugin()
	err = mssqlPlugin.Init()
	if err != nil {
		logger.Log1.Errorf("初始化 MSSQL 插件失败: %v", err)
	}
	pm.RegisterPlugin(mssqlPlugin.Name, mssqlPlugin)

	// 5. pg sql 插件
	pgsqlPlugin := NewPGSQLPlugin()
	err = pgsqlPlugin.Init()
	if err != nil {
		logger.Log1.Errorf("初始化 PGSQL 插件失败: %v", err)
	}
	pm.RegisterPlugin(pgsqlPlugin.Name, pgsqlPlugin)

	// 6. oracle db 插件
	oracleDBPlugin := NewOracleDBPlugin()
	err = oracleDBPlugin.Init()
	if err != nil {
		logger.Log1.Errorf("初始化 OracleDB 插件失败: %v", err)
	}
	pm.RegisterPlugin(oracleDBPlugin.Name, oracleDBPlugin)

	return nil
}

func (pm *PluginManager) RegisterPlugin(name string, plugin Plugin) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.plugins[name] = plugin
	logger.Log1.WithField("plugin", name).Info("插件已注册")
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

func (pm *PluginManager) CloseAll() {
	if len(pm.plugins) > 0 {
		pm.mu.RLock()
		defer pm.mu.RUnlock()
		for name, plugin := range pm.plugins {
			err := plugin.Close()
			if err != nil {
				logger.Log1.Errorf("关闭插件 %s 失败: %v", name, err)
			}
		}
	}
}
