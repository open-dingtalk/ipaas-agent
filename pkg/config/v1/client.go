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
}

func (c *ClientCommonConfig) Complete() {
	for i := range c.Plugins {
		c.Plugins[i].Complete()
	}

	if c.Auth.OpenAPIHost == "" {
		c.Auth.OpenAPIHost = "https://api.dingtalk.com"
	}
}
