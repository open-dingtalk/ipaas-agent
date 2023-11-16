package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func HandleMySQLProxyRequesr(agentProtocol *IPaaSAgentProtocol) ([]byte, error) {
	logger := zap.L()
	mySqlConfig := config.MySQL[0]
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", mySqlConfig.Username, mySqlConfig.Password, mySqlConfig.Addr, mySqlConfig.Database))
	if err != nil {
		panic(err)
	}
	// See "Important settings" section.
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Query
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		panic(err)
	}
	columns, err := rows.Columns()
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	err = rows.Err()
	if err != nil {
		panic(err)
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
