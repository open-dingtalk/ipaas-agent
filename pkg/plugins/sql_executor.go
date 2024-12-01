package plugins

import (
	"database/sql"
	"encoding/json"
	"strconv"
)

type SQLExecutor interface {
	GetConnection(body *Body) (*sql.DB, error)
	DoSQLExecute(body *Body) *QueryResult
}

type FlexInt int

func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	// 尝试作为数字解析
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*fi = FlexInt(i)
		return nil
	}

	// 尝试作为字符串解析
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	// 将字符串转换为整数
	v, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	*fi = FlexInt(v)
	return nil
}

// Body 结构体定义数据库连接参数
type Body struct {
	Host     string  `json:"host,omitempty" mapstructure:"host,omitempty"`
	Port     FlexInt `json:"port,omitempty" mapstructure:"port,omitempty"`
	User     string  `json:"user,omitempty" mapstructure:"user,omitempty"`
	Password string  `json:"password,omitempty" mapstructure:"password,omitempty"`
	Database string  `json:"database,omitempty" mapstructure:"database,omitempty"`
	SQL      string  `json:"sql,omitempty" mapstructure:"sql,omitempty"`
	// 以下字段用于本地网关配置
	Address    string `json:"address,omitempty" mapstructure:"address,omitempty"`
	ConfigKey  string `json:"config_key,omitempty" mapstructure:"config_key,omitempty"`
	ConnString string `json:"connection_str,omitempty" mapstructure:"connection_str,omitempty"`
}

// 从本地配置中完善 Body
func (b *Body) completeFrom(other *Body) {
	b.Host = other.Host
	b.Port = other.Port
	b.User = other.User
	b.Password = other.Password
	b.Database = other.Database
	if other.SQL != "" {
		b.SQL = other.SQL
	}
	b.Address = other.Address
	b.ConfigKey = other.ConfigKey
	b.ConnString = other.ConnString
}

// QueryResult 结构体定义查询结果
type QueryResult struct {
	Result  []map[string]interface{} `json:"result" mapstructure:"result,omitempty"`
	Columns []string                 `json:"columns" mapstructure:"columns,omitempty"`
	Message string                   `json:"message" mapstructure:"message,omitempty"`
}
