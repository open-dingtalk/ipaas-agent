package v1

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type ClientPluginOptions interface {
	Complete()
}

type TypedClientPluginOptions struct {
	Type string `json:"type"`
	ClientPluginOptions
}

func (c *TypedClientPluginOptions) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if rawType, ok := raw["type"].(string); ok {
		c.Type = rawType
	}
	if c.Type == "" {
		return fmt.Errorf("missing type field")
	}
	if t, ok := clientPluginOptionsTypeMap[c.Type]; ok {
		c.ClientPluginOptions = reflect.New(t).Interface().(ClientPluginOptions)
	}
	return json.Unmarshal(data, c.ClientPluginOptions)
}

const (
	PluginMySQL = "mysql"
)

var clientPluginOptionsTypeMap = map[string]reflect.Type{
	PluginMySQL: reflect.TypeOf(MySQLPluginOptions{}),
}

type MySQLPluginOptions struct {
	Host      string `json:"host,omitempty"`
	Port      int    `json:"port,omitempty"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	Database  string `json:"database,omitempty"`
	ConfigKey string `json:"configKey,omitempty"`
}

func (o *MySQLPluginOptions) Complete() {
	if o.Port == 0 {
		o.Port = 3306
	}
	if o.ConfigKey == "" {
		o.ConfigKey = "default"
	}
	if o.Host == "" {
		o.Host = "127.0.0.1"
	}
}
