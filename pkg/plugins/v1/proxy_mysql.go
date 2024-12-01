package v1

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
)

// see com.dingtalk.open.connect.engine.support.utils.IPaaSAgentUtils.Body
type MySQLAgentProtocol struct {
	ConfigKey    string       `json:"configKey"`
	ConfigParams ConfigParams `json:"configParams"`
}

type ConfigParams struct {
	Sql string `json:"sql"`
}

type IPaaSAgentProtocol struct {
	Headers Headers `json:"headers"`
	Body    Body    `json:"body"`
}

type Headers struct {
	SpecVersion     string `json:"specVersion"`
	ConnectorCorpId string `json:"connectorCorpId"`
	Type            string `json:"type"`
	// connector property
	ConnectorId string `json:"connectorId"`
	ActionId    string `json:"actionId"`
}

type Body struct {
	HTTPRequest  HTTPRequest       `json:"httpRequest"`
	ConfigParams map[string]string `json:"configParams"`
	ConfigKey    string            `json:"configKey"`
}

type MySQLConfig struct {
	Addr      string `mapstructure:"addr,omitempty" json:"addr,omitempty" `
	Username  string `mapstructure:"username,omitempty" json:"username,omitempty" `
	Password  string `mapstructure:"password,omitempty" json:"password,omitempty" `
	Database  string `mapstructure:"database,omitempty" json:"database,omitempty" `
	ConfigKey string `mapstructure:"config_key,omitempty" json:"config_key,omitempty" `
}

func HandleMySQLProxyRequest(agentProtocol *IPaaSAgentProtocol) (interface{}, error) {
	logger.Log1.Infof("handle mysql proxy request: %v", agentProtocol)
	mysqlProtocol := &MySQLAgentProtocol{
		ConfigKey:    agentProtocol.Body.ConfigKey,
		ConfigParams: ConfigParams{Sql: agentProtocol.Body.ConfigParams["sql"]},
	}
	logger.Log1.Infof("mysql protocol: %v", mysqlProtocol)

	mySqlConfig := findConfigByKey(mysqlProtocol.ConfigKey)
	if mySqlConfig == nil {
		logger.Log1.Errorf("mysql config %s not found", mysqlProtocol.ConfigKey)
		return nil, nil
	}
	logger.Log1.Infof("mysql config: %v", mySqlConfig)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", mySqlConfig.Username, mySqlConfig.Password, mySqlConfig.Addr, mySqlConfig.Database))
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Query
	runSql := mysqlProtocol.ConfigParams.Sql
	rows, err := db.Query(runSql)
	if err != nil {
		logger.Log1.Errorf("mysql query error: %v", err)
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		logger.Log1.Errorf("mysql query error: %v", err)
		return nil, err
	}

	defer rows.Close()
	err = rows.Err()
	if err != nil {
		logger.Log1.Errorf("mysql query error: %v", err)
		return nil, err
	}
	response := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scanArgs := make([]interface{}, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		err = rows.Scan(scanArgs...)
		if err != nil {
			logger.Log1.Errorf("mysql query error: %v", err)
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range values {
			if value, ok := col.([]byte); ok {
				row[columns[i]] = string(value)
			} else {
				row[columns[i]] = col
			}
		}
		logger.Log1.Infof("mysql query result: %v", row)
		response = append(response, row)
	}

	db.Close()
	return response, nil
}

func findConfigByKey(key string) *MySQLConfig {
	// 定义一个变量来存储 MySQL 配置
	var mysqlConfigs []MySQLConfig

	// 解析 MySQL 配置
	if err := viper.UnmarshalKey("mysql", &mysqlConfigs); err != nil {
		logger.Log1.Errorf("解析 MySQL 配置出错: %v", err)
	}
	// 打印 MySQL 配置
	for _, config := range mysqlConfigs {
		if config.ConfigKey == key {
			return &config
		}
	}
	return nil
}
