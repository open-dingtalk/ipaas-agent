package plugins

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
	go_ora "github.com/sijms/go-ora/v2"
	"github.com/spf13/viper"
)

type OracleDBPlugin struct {
	Name        string
	AllowRemote bool
	Configs     []Body
}

func (p *OracleDBPlugin) GetConnection(body *Body) (*sql.DB, error) {
	var urlOptions map[string]string
	if body.SID != "" {
		urlOptions = map[string]string{
			"SID": body.SID,
		}
	}
	connString := go_ora.BuildUrl(
		body.Host, int(body.Port), body.ServiceName, body.User, body.Password, urlOptions,
	)

	return sql.Open("oracle", connString)
}

// doSQLExecute 执行SQL查询
func (p *OracleDBPlugin) DoSQLExecute(body *Body) (qr *QueryResult) {
	startTime := time.Now()
	defer func() {
		if qr != nil && qr.Message != "success" {
			logger.Log1.WithField("cost", time.Since(startTime).String()).Errorf("SQL查询结束")
		} else {
			logger.Log1.WithField("cost", time.Since(startTime).String()).Infof("SQL查询结束")
		}
	}()

	// 获取数据库连接
	db, err := p.GetConnection(body)
	if err != nil {
		logger.Log1.WithField("error", err).Error("获取数据库连接失败")
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
		logger.Log1.WithField("error", err).Error("执行SQL失败")
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
		logger.Log1.WithField("error", err).Error("获取列名失败")
		return &QueryResult{
			Result:  nil,
			Columns: nil,
			Message: err.Error(),
		}
	}

	// 获取列的类型信息
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		logger.Log1.WithField("error", err).Error("获取列类型失败")
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
		for i, colType := range columnTypes {
			// 根据列的扫描类型创建对应的变量
			values[i] = reflect.New(colType.ScanType()).Interface()
		}

		// 扫描行数据
		err := rows.Scan(values...)
		if err != nil {
			continue
		}

		// 将行数据转换为 map
		row := make(map[string]interface{})
		for i, col := range columns {
			// 处理指针类型，获取实际的值
			val := values[i]
			if bv, ok := val.(*interface{}); ok {
				row[col] = *bv
			} else {
				row[col] = val
			}
		}
		result = append(result, row)
	}

	return &QueryResult{
		Result:  result,
		Columns: columns,
		Message: "success",
	}
}

func (p *OracleDBPlugin) findConfigByKey(key string) *Body {
	for _, config := range p.Configs {
		if config.ConfigKey == key {
			logger.Log1.WithField("config", config).Info("找到配置")
			return &config
		}
	}
	return nil
}

func NewOracleDBPlugin() *OracleDBPlugin {
	return &OracleDBPlugin{
		Name: "oracledb_plugin",
	}
}

func (p *OracleDBPlugin) Init() error {
	// 定义一个变量来存储 SQL 配置
	var sqlConfigs []Body

	// 解析 SQL 配置
	if err := viper.UnmarshalKey("plugins.oracledb", &sqlConfigs); err != nil {
		logger.Log1.Fatalf("解析 oracle 数据库配置出错: %v", err)
	}

	p.Configs = sqlConfigs

	p.AllowRemote = viper.GetBool("auth.oracledb.allow_remote")

	logger.Log1.
		WithField("插件名", p.Name).
		WithField("配置列表", p.Configs).
		WithField("允许远程配置", p.AllowRemote).
		Info("插件已初始化")
	return nil
}

func (p *OracleDBPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
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

func (p *OracleDBPlugin) Close() error {
	// 关闭插件
	logger.Log1.WithField("plugin", p.Name).Info("插件已关闭")
	return nil
}
