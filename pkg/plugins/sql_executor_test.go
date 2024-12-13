package plugins_test

import (
	"encoding/json"
	"testing"

	plugin "github.com/open-dingtalk/ipaas-agent/pkg/plugins"
	"github.com/stretchr/testify/require"
)

func TestPGSQLPlugin_doSQLExecute(t *testing.T) {
	// 创建一个 MSSQL 插件
	p := &plugin.PGSQLPlugin{
		Name:        "",
		AllowRemote: true,
	}
	// 创建一个 Body
	body := &plugin.Body{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "example",
		Database: "example",
		SQL:      "SELECT * FROM \"user\" LIMIT 50",
	}
	// 执行 SQL 查询
	qr := p.DoSQLExecute(body)
	// 断言结果
	require.NotNil(t, qr)
	require.NotNil(t, qr.Result)
	require.NotNil(t, qr.Columns)
	// 打印到控制台
	for _, row := range qr.Result {
		for key, col := range row {
			switch v := col.(type) {
			case []byte:
				t.Logf("%s: %s", key, string(v))
			default:
				t.Logf("%s: %v", key, v)
			}
		}
	}

	require.Equal(t, "success", qr.Message)
}

func TestMYSQLPlugin_doSQLExecute(t *testing.T) {
	// 创建一个 MSSQL 插件
	p := &plugin.MySQLPlugin{
		Name:        "",
		AllowRemote: true,
	}
	// 创建一个 Body
	body := &plugin.Body{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "root",
		Database: "example",
		SQL:      "SELECT * FROM `users` LIMIT 50",
	}
	// 执行 SQL 查询
	qr := p.DoSQLExecute(body)
	// 断言结果
	require.NotNil(t, qr)
	require.NotNil(t, qr.Result)
	require.NotNil(t, qr.Columns)
	// 打印到控制台
	for _, row := range qr.Result {
		for key, col := range row {
			switch v := col.(type) {
			case []byte:
				t.Logf("%s: %s", key, string(v))
			default:
				t.Logf("%s: %v", key, v)
			}
		}
	}

	require.Equal(t, "success", qr.Message)
}

func TestORACLEDBPlugin_doSQLExecute(t *testing.T) {
	// 创建一个 MSSQL 插件
	p := &plugin.OracleDBPlugin{
		Name:        "",
		AllowRemote: true,
	}
	// 创建一个 Body
	body := &plugin.Body{
		Host:     "localhost",
		Port:     1521,
		User:     "system",
		Password: "example",
		SID:      "FREE",
		SQL:      "SELECT * FROM HELP WHERE ROWNUM <= 10",
	}
	// 执行 SQL 查询
	qr := p.DoSQLExecute(body)
	// 断言结果
	require.NotNil(t, qr)
	require.NotNil(t, qr.Result)
	require.NotNil(t, qr.Columns)
	// 打印到控制台
	for _, row := range qr.Result {
		for key, col := range row {
			switch v := col.(type) {
			case []byte:
				t.Logf("%s: %s", key, string(v))
			default:
				t.Logf("%s: %v", key, v)
			}
		}
	}

	require.Equal(t, "success", qr.Message)
}

func TestMSSQLPlugin_doSQLExecute(t *testing.T) {
	// 创建一个 MSSQL 插件
	p := &plugin.MSSQLPlugin{
		Name:                 "",
		AllowRemote:          true,
		LessCommonParameters: "encrypt=disable;trustServerCertificate=true",
	}
	// 创建一个 Body
	body := &plugin.Body{
		Host:     "localhost",
		Port:     1433,
		User:     "sa",
		Password: "sa123456A",
		Database: "TestDB",
		SQL:      "SELECT * FROM Employees",
	}
	// 执行 SQL 查询
	qr := p.DoSQLExecute(body)
	// 断言结果
	require.NotNil(t, qr)
	require.NotNil(t, qr.Result)
	require.NotNil(t, qr.Columns)
	jsonData, err := json.Marshal(qr.Result)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(jsonData))

	require.Equal(t, "success", qr.Message)
}

func TestMSSQLPlugin_doSQLExecute2(t *testing.T) {
	// 创建一个 MSSQL 插件
	p := &plugin.MSSQLPlugin{
		Name:                 "",
		AllowRemote:          true,
		LessCommonParameters: "encrypt=disable;trustServerCertificate=true",
	}
	// 创建一个 Body
	body := &plugin.Body{
		Host:     "localhost",
		Port:     1433,
		User:     "sa",
		Password: "sa123456A",
		Database: "master",
		SQL:      "SELECT * FROM AllDataTypesTest",
	}
	// 执行 SQL 查询
	qr := p.DoSQLExecute(body)
	// 断言结果
	require.NotNil(t, qr)
	require.NotNil(t, qr.Result)
	require.NotNil(t, qr.Columns)
	jsonData, err := json.Marshal(qr.Result)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(jsonData))

	require.Equal(t, "success", qr.Message)
}
