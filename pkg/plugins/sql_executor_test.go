package plugins_test

import (
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	plugin "github.com/open-dingtalk/ipaas-agent/pkg/plugins"
	"github.com/stretchr/testify/require"
)

// requireLiveDB 保护需要真实数据库实例的用例。
//
// 下面几个用例依赖按 README 手工搭建的实例，以及固定的库名、账号和表结构，
// 在干净的克隆或 CI 上没有这些前提，因此默认跳过；设置了 IPAAS_AGENT_DB_TEST=1
// 才真正执行，并且会先确认端口可达，避免把"没起数据库"报成代码问题。
func requireLiveDB(t *testing.T, addr string) {
	t.Helper()
	if os.Getenv("IPAAS_AGENT_DB_TEST") != "1" {
		t.Skip("该用例需要真实数据库实例，设置 IPAAS_AGENT_DB_TEST=1 后运行")
	}
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		t.Skipf("%s 不可达，跳过: %v", addr, err)
	}
	require.NoError(t, conn.Close())
}

func TestPGSQLPlugin_doSQLExecute(t *testing.T) {
	requireLiveDB(t, "localhost:5432")
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
	requireLiveDB(t, "localhost:3306")
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
	requireLiveDB(t, "localhost:1521")
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
	requireLiveDB(t, "localhost:1433")
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
	requireLiveDB(t, "localhost:1433")
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
