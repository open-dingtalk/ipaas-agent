package config

import (
	"testing"

	"os"
	"time"

	v1 "github.com/open-dingtalk/ipaas-agent/pkg/config/v1"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// 创建一个临时配置文件
	configContent := `
auth:
  clientID: "clientID111"
  clientSecret: "clientSecret"
  openAPIHost: "openAPIHost"
plugins:
  - type: "mysql"
    host: "localhost"
    port: 3306
    username: "root"
    password: "root"
    database: "example"
    configKey: "default"
  - type: "mysql"
    host: "localhost"
    port: 3307
    username: "root"
    password: "root"
    database: "example"
    configKey: "default2"
`
	configFile, err := os.CreateTemp("", "config.yaml")
	require.NoError(t, err)
	defer os.Remove(configFile.Name())

	_, err = configFile.Write([]byte(configContent))
	require.NoError(t, err)
	configFile.Close()

	// 设置 Viper 读取临时配置文件
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile.Name())

	// 打印配置文件内容
	err = viper.ReadInConfig()
	require.NoError(t, err)

	configMap := viper.AllSettings()
	for key, value := range configMap {
		t.Logf("%s: %v\n", key, value)
	}

	// 调用 LoadConfig 函数
	LoadConfig()
	auth := GetAuthClientConfig()
	require.NoError(t, err)

	// 验证配置内容
	require.Equal(t, "clientID111", auth.ClientID)
	require.Equal(t, "clientSecret", auth.ClientSecret)
	require.Equal(t, "openAPIHost", auth.OpenAPIHost)

	// require.Len(t, config.Plugins, 2)
	// require.Equal(t, "mysql", config.Plugins[0].Type)
	// require.Equal(t, "localhost", config.Plugins[0].ClientPluginOptions.(*v1.MySQLPluginOptions).Host)
	// require.Equal(t, 3306, config.Plugins[0].ClientPluginOptions.(*v1.MySQLPluginOptions).Port)
	// require.Equal(t, "root", config.Plugins[0].ClientPluginOptions.(*v1.MySQLPluginOptions).Username)
	// require.Equal(t, "root", config.Plugins[0].ClientPluginOptions.(*v1.MySQLPluginOptions).Password)
	// require.Equal(t, "example", config.Plugins[0].ClientPluginOptions.(*v1.MySQLPluginOptions).Database)
	// require.Equal(t, "default", config.Plugins[0].ClientPluginOptions.(*v1.MySQLPluginOptions).ConfigKey)

	// require.Equal(t, "mysql", config.Plugins[1].Type)
	// require.Equal(t, "localhost", config.Plugins[1].ClientPluginOptions.(*v1.MySQLPluginOptions).Host)
	// require.Equal(t, 3307, config.Plugins[1].ClientPluginOptions.(*v1.MySQLPluginOptions).Port)
	// require.Equal(t, "root", config.Plugins[1].ClientPluginOptions.(*v1.MySQLPluginOptions).Username)
	// require.Equal(t, "root", config.Plugins[1].ClientPluginOptions.(*v1.MySQLPluginOptions).Password)
	// require.Equal(t, "example", config.Plugins[1].ClientPluginOptions.(*v1.MySQLPluginOptions).Database)
	// require.Equal(t, "default2", config.Plugins[1].ClientPluginOptions.(*v1.MySQLPluginOptions).ConfigKey)
}

// TestGetTimeoutConfig 校验超时兜底策略：平台没下发值时不能退化成"永不超时"，
// 下发了值时按平台优先、只在超过本地上限时截断。
func TestGetTimeoutConfig(t *testing.T) {
	t.Cleanup(func() {
		viper.Set("timeout.http_default_ms", 0)
		viper.Set("timeout.http_max_ms", 0)
		viper.Set("timeout.db_ms", 0)
	})

	// 未配置：回落到内置默认值
	viper.Set("timeout.http_default_ms", 0)
	viper.Set("timeout.http_max_ms", 0)
	viper.Set("timeout.db_ms", 0)
	conf := GetTimeoutConfig()
	require.Equal(t, v1.DefaultHTTPTimeoutMS, conf.HTTPDefaultMS)
	require.Equal(t, v1.DefaultDBTimeoutMS, conf.DBMS)
	require.Equal(t, time.Duration(v1.DefaultHTTPTimeoutMS)*time.Millisecond, conf.HTTPTimeout(0))

	// 显式配置：按配置生效
	viper.Set("timeout.http_default_ms", 8000)
	viper.Set("timeout.http_max_ms", 10000)
	viper.Set("timeout.db_ms", 1500)
	conf = GetTimeoutConfig()
	require.Equal(t, 1500*time.Millisecond, conf.DBTimeout())
	require.Equal(t, 8*time.Second, conf.HTTPTimeout(0), "平台未下发时使用本地默认值")
	require.Equal(t, 9*time.Second, conf.HTTPTimeout(9000), "平台下发的值优先")
	require.Equal(t, 10*time.Second, conf.HTTPTimeout(60000), "超过本地上限时截断")

	// 负数上限按未配置处理
	viper.Set("timeout.http_max_ms", -1)
	conf = GetTimeoutConfig()
	require.Equal(t, 60*time.Second, conf.HTTPTimeout(60000))
}

// TestGetAuthClientConfigProxy 校验代理配置的两个入口：auth.proxy 优先，client.proxy 兼容旧写法。
func TestGetAuthClientConfigProxy(t *testing.T) {
	t.Cleanup(func() {
		viper.Set("auth.proxy", "")
		viper.Set("client.proxy", "")
	})

	viper.Set("auth.clientID", "id")
	viper.Set("auth.clientSecret", "secret")
	viper.Set("auth.proxy", "")
	viper.Set("client.proxy", "http://127.0.0.1:7890")
	require.Equal(t, "http://127.0.0.1:7890", GetAuthClientConfig().Proxy)

	viper.Set("auth.proxy", "http://10.0.0.1:8080")
	require.Equal(t, "http://10.0.0.1:8080", GetAuthClientConfig().Proxy)

	viper.Set("auth.proxy", "")
	viper.Set("client.proxy", "")
	require.Empty(t, GetAuthClientConfig().Proxy, "未配置时保持直连")
}
