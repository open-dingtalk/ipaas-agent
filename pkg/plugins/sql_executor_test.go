package plugins_test

import (
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
