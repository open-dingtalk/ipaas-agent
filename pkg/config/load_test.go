package config

import (
	"testing"

	v1 "github.com/open-dingtalk/ipaas-agent/pkg/config/v1"
	"github.com/stretchr/testify/require"
)

const jsonClientContent = `
{
  "auth": {
    "clientID": "clientID111",
	"clientSecret": "clientSecret"
	},
	"plugins": [{
		"type": "mysql",
		"host": "localhost",
		"port": 3306,
		"username": "root",
		"password": "root",
		"database": "example",
		"configKey": "default"
	}, {
		"type": "mysql",
		"host": "localhost",
		"port": 3307,
		"username": "root",
		"password": "root",
		"database": "example",
		"configKey": "default2"
	}]
}
`

const yamlClientContent = `
auth:
  clientID: clientID111
  clientSecret: clientSecret
plugins:
  - type: mysql
    host: localhost
    port: 3306
    username: root
    password: root
    database: example
    configKey: default
  - type: mysql
    host: localhost
    port: 3307
    username: root
    password: root
    database: example
    configKey: default2
`

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"json", jsonClientContent},
		{"yaml", yamlClientContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)
			c := v1.ClientConfig{}
			err := LoadConfigure([]byte(tt.content), &c, true)
			if err != nil {
				t.Errorf("LoadConfigure() error = %v", err)
			}
			require.NoError(err)
			require.EqualValues("clientID111", c.Auth.ClientID)
			require.EqualValues("clientSecret", c.Auth.ClientSecret)
			require.Len(c.Plugins, 2)
			require.EqualValues("mysql", c.Plugins[0].Type)
		})
	}
}
