package v1

// type ClientConfig struct {
// 	ClientCommonConfig
// }

type ClientCommonConfig struct {
	Auth    AuthClientConfig           `json:"auth"`
	Proxies ProxyBaseConfig            `json:"proxies,omitempty"`
	Plugins []TypedClientPluginOptions `json:"plugins,omitempty"`
}

type AuthClientConfig struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	OpenAPIHost  string `json:"openAPIHost"`
	// Proxy 连接钉钉开放平台使用的代理地址（如 http://127.0.0.1:7890），
	// 留空表示直连。内网只允许走代理出网时必须配置。
	Proxy string `json:"proxy,omitempty"`
}

func (c *ClientCommonConfig) Complete() {
	for i := range c.Plugins {
		c.Plugins[i].Complete()
	}

	if c.Auth.OpenAPIHost == "" {
		c.Auth.OpenAPIHost = "https://api.dingtalk.com"
	}
}
