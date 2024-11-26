package config

import (
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/config/v1"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	"github.com/spf13/viper"
)

var (
	glbEnvs map[string]string
	mu      sync.RWMutex
)

func init() {
	glbEnvs = make(map[string]string)
	envs := os.Environ()
	for _, env := range envs {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		glbEnvs[pair[0]] = pair[1]
	}
}

func LoadConfig() error {
	mu.Lock()
	defer mu.Unlock()

	logger.Log1.Info("加载配置文件...")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// 添加配置文件路径
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	err := viper.ReadInConfig()
	if err != nil {
		logger.Log1.Errorf("读取配置文件出错: %v", err)
		return err
	}

	// 启用从环境变量读取配置
	viper.AutomaticEnv()

	// 从环境变量中读取配置
	viper.BindEnv("auth.openAPIHost", "IPAAS_AGENT_AUTH_OPEN_API_HOST")

	return nil
}

func WatchConfig(onChange func()) {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		mu.Lock()
		defer mu.Unlock()
		logger.Log1.Infof("配置文件发生变化: %s", e.Name)
		onChange()
	})
}

func GetAuthClientConfig() *v1.AuthClientConfig {
	mu.RLock()
	defer mu.RUnlock()
	auth := &v1.AuthClientConfig{
		ClientID: FirstNonEmpty(
			viper.GetString("auth.clientID"),
			viper.GetString("client.client_id"), // 兼容旧版本
		),
		ClientSecret: FirstNonEmpty(
			viper.GetString("auth.clientSecret"),
			viper.GetString("client.client_secret"), // 兼容旧版本
		),
		OpenAPIHost: FirstNonEmpty(
			viper.GetString("auth.openAPIHost"),
			"https://api.dingtalk.com",
		),
	}
	return auth
}

func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
