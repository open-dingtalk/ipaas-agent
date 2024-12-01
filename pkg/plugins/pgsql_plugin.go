package plugins

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	_ "github.com/lib/pq"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
	"github.com/spf13/viper"
)

type PGSQLPlugin struct {
	Name        string
	AllowRemote bool
	Configs     []Body
}

func (p *PGSQLPlugin) GetConnection(body *Body) (*sql.DB, error) {
	connString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		body.Host, body.Port, body.User, body.Password, body.Database)

	return sql.Open("postgres", connString)
}

// doSQLExecute 执行SQL查询
func (p *PGSQLPlugin) DoSQLExecute(body *Body) (qr *QueryResult) {
	startTime := time.Now()
	defer func() {
		logger.Log1.WithField("cost", time.Since(startTime).String()).Infof("SQL查询结束")
	}()

	// 获取数据库连接
	db, err := p.GetConnection(body)
	if err != nil {
		return &QueryResult{
			Result:  nil,
			Columns: nil,
			Message: err.Error(),
		}
	}
	defer db.Close()

	// Sleep for 10000ms to simulate processing time
	// time.Sleep(10000 * time.Millisecond)

	rows, err := db.Query(body.SQL)
	if err != nil {
		return &QueryResult{
			Result:  nil,
			Columns: nil,
			Message: err.Error(),
		}
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return &QueryResult{
			Result:  nil,
			Columns: nil,
			Message: err.Error(),
		}
	}

	// 准备结果集
	var result []map[string]interface{}

	// 扫描每一行
	for rows.Next() {
		// 创建一个切片，用于存储一行的值
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		err := rows.Scan(values...)
		if err != nil {
			continue
		}

		// 将行数据转换为map
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i].(*interface{})
			row[col] = *val
		}
		result = append(result, row)
	}

	return &QueryResult{
		Result:  result,
		Columns: columns,
		Message: "success",
	}
}

func (p *PGSQLPlugin) findConfigByKey(key string) *Body {
	for _, config := range p.Configs {
		if config.ConfigKey == key {
			logger.Log1.WithField("config", config).Info("找到配置")
			return &config
		}
	}
	return nil
}

func NewPGSQLPlugin() *PGSQLPlugin {
	return &PGSQLPlugin{
		Name: "pgsql_plugin",
	}
}

func (p *PGSQLPlugin) Init() error {
	// 定义一个变量来存储 SQL 配置
	var sqlConfigs []Body

	// 解析 SQL 配置
	if err := viper.UnmarshalKey("plugins.mssql", &sqlConfigs); err != nil {
		logger.Log1.Fatalf("解析 MSSQL 配置出错: %v", err)
	}

	p.Configs = sqlConfigs

	p.AllowRemote = viper.GetBool("auth.mssql.allow_remote")

	logger.Log1.
		WithField("插件名", p.Name).
		WithField("配置列表", p.Configs).
		WithField("允许远程配置", p.AllowRemote).
		Info("插件已初始化")
	return nil
}

func (p *PGSQLPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
	// 初始化 Data
	data, err := df.GetPluginDataWithType(reflect.TypeOf(Body{}))

	if err != nil {
		return payload.NewErrorDataFrameResponse(err), err
	}

	remoteConf := data.(*Body)
	if remoteConf.ConfigKey == "" && p.AllowRemote {
		logger.Log1.WithField("config", remoteConf).Info("使用远程配置")
	} else {
		localConf := p.findConfigByKey(remoteConf.ConfigKey)
		if localConf == nil {
			logger.Log1.WithField("configKey", remoteConf.ConfigKey).
				WithField("是否允许远程配置", p.AllowRemote).
				Error("未找到配置或不允许远程配置")
			return payload.NewErrorDataFrameResponse(fmt.Errorf("未找到配置或不允许远程配置: %s", remoteConf.ConfigKey)), nil
		}
		remoteConf.completeFrom(localConf)
	}

	callBackResponse := &CallbackResponse{
		Response: p.DoSQLExecute(remoteConf),
	}

	resp := payload.NewSuccessDataFrameResponse()

	resp.SetJson(callBackResponse)

	return resp, nil
}

func (p *PGSQLPlugin) Close() error {
	// 关闭插件
	logger.Log1.WithField("plugin", p.Name).Info("插件已关闭")
	return nil
}
