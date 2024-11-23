package v1

type ClientConfig struct {
	ClientCommonConfig
}

type ClientCommonConfig struct {
	Auth    AuthClientConfig           `json:"auth"`
	Proxies ProxyBaseConfig            `json:"proxies,omitempty"`
	Plugins []TypedClientPluginOptions `json:"plugins,omitempty"`
}

type AuthClientConfig struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

func (c *ClientConfig) Complete() {
	for i := range c.Plugins {
		c.Plugins[i].Complete()
	}
}
