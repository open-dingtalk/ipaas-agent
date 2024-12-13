package plugins

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/microsoft/go-mssqldb/integratedauth/krb5"
	"github.com/spf13/viper"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

// MYSQLPlugin 结构体定义 MySQL 插件
type MSSQLPlugin struct {
	Name        string
	AllowRemote bool
	Configs     []Body
}

// getConnection 创建数据库连接
func (p *MSSQLPlugin) GetConnection(body *Body) (*sql.DB, error) {
	connString := fmt.Sprintf("server=%s;port=%d;user id=%s;password=%s;database=%s",
		body.Host, body.Port, body.User, body.Password, body.Database)

	return sql.Open("sqlserver", connString)
}

// doSQLExecute 执行SQL查询
func (p *MSSQLPlugin) DoSQLExecute(body *Body) (qr *QueryResult) {
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

	rows, err := db.Query(body.SQL)
	if err != nil {
		logger.Log1.WithField("error", err).Error("SQL查询失败")
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
	rowCount := 0

	// 扫描每一行
	for rows.Next() {
		rowCount++
		logger.Log1.WithField("rowCount", rowCount).Debugf("正在处理第 %d 行数据", rowCount)
		// 创建一个切片，用于存储一行的值
		values := make([]interface{}, len(columns))
		for i, colType := range columnTypes {
			// 根据列的扫描类型创建对应的变量
			values[i] = reflect.New(colType.ScanType()).Interface()
			dbType := colType.DatabaseTypeName()
			// 根据数据库类型创建对应的变量
			switch dbType {
			case "DECIMAL", "NUMERIC", "FLOAT", "REAL":
				var v float64
				values[i] = &v
			case "BIGINT", "INT", "SMALLINT", "TINYINT":
				var v int64
				values[i] = &v
			case "BIT":
				var v bool
				values[i] = &v
			case "DATETIME", "DATETIME2", "DATE", "TIME":
				var v time.Time
				values[i] = &v
			default: // VARCHAR, NVARCHAR, CHAR, NCHAR, TEXT 等
				var v string
				values[i] = &v
			}
		}

		// 扫描行数据
		err := rows.Scan(values...)
		if err != nil {
			logger.Log1.WithFields(map[string]interface{}{
				"error": err,
				"row":   rowCount,
			}).Error("扫描行数据失败, 跳过")
			continue
		}

		// 将行数据转换为 map
		row := make(map[string]interface{})
		for i, col := range columns {
			// 处理指针类型，获取实际的值
			// row[col] = *(values[i].(*interface{}))
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

func (p *MSSQLPlugin) findConfigByKey(key string) *Body {
	for _, config := range p.Configs {
		if config.ConfigKey == key {
			logger.Log1.WithField("config", config).Info("找到配置")
			return &config
		}
	}
	return nil
}

func NewMSSQLPlugin() *MSSQLPlugin {
	return &MSSQLPlugin{
		Name: "mssql_plugin",
	}
}

func (p *MSSQLPlugin) Init() error {
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

func (p *MSSQLPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
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

func (p *MSSQLPlugin) Close() error {
	// 关闭插件
	logger.Log1.WithField("plugin", p.Name).Info("插件已关闭")
	return nil
}
