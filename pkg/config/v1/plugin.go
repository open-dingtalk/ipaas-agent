package v1

import (
	"reflect"
)

type ClientPluginOptions interface {
	Complete()
}

type TypedClientPluginOptions struct {
	Type string `json:"type"`
	ClientPluginOptions
}

const (
	PluginMySQL = "mysql"
)

var ClientPluginOptionsTypeMap = map[string]reflect.Type{
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
