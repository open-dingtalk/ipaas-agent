package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"github.com/open-dingtalk/ipaas-net-gateway/config"
)

// see com.dingtalk.open.connect.engine.support.utils.IPaaSAgentUtils.Body
type MySQLAgentProtocol struct {
	ConfigKey    string       `json:"configKey"`
	ConfigParams ConfigParams `json:"configParams"`
}

type ConfigParams struct {
	Sql string `json:"sql"`
}

func HandleMySQLProxyRequest(agentProtocol *IPaaSAgentProtocol) ([]byte, error) {
	logger := zap.L()
	logger.Info("handle mysql proxy request", zap.Any("agentProtocol", agentProtocol))
	mysqlProtocol := &MySQLAgentProtocol{
		ConfigKey:    agentProtocol.Body.ConfigKey,
		ConfigParams: ConfigParams{Sql: agentProtocol.Body.ConfigParams["sql"]},
	}
	logger.Info("mysql protocol", zap.Any("mysqlProtocol", mysqlProtocol))

	mySqlConfig := findConfigByKey(mysqlProtocol.ConfigKey)
	if mySqlConfig == nil {
		logger.Error("mysql config not found", zap.String("configId", mysqlProtocol.ConfigKey))
		return nil, nil
	}
	logger.Info("mysql config", zap.Any("mySqlConfig", mySqlConfig))
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
		logger.Error("mysql query error", zap.Error(err))
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		logger.Error("mysql query error", zap.Error(err))
		return nil, err
	}

	defer rows.Close()
	err = rows.Err()
	if err != nil {
		logger.Error("mysql query error", zap.Error(err))
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
			panic(err)
		}

		row := make(map[string]interface{})
		for i, col := range values {
			if value, ok := col.([]byte); ok {
				row[columns[i]] = string(value)
			} else {
				row[columns[i]] = col
			}
		}
		logger.Info("mysql query result", zap.Any("row", row))
		response = append(response, row)
	}

	db.Close()
	return json.Marshal(response)
}

func findConfigByKey(key string) *config.MySqlConfig {
	configs := config.GetConfig().MySQL
	for _, config := range configs {
		if config.ConfigKey == key {
			return config
		}
	}
	return nil
}
