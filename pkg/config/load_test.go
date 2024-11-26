package config

import (
	"testing"

	"os"

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
