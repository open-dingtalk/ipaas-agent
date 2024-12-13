package plugins

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
	v1 "github.com/open-dingtalk/ipaas-agent/pkg/plugins/v1"
)

// MYSQLPlugin 结构体定义 MySQL 插件
type MySQLPlugin struct {
	Name         string
	AllowRemote  bool
	ValueAsBytes bool
	Configs      []Body
}

// getConnection 创建数据库连接
func (p *MySQLPlugin) GetConnection(body *Body) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		body.User,
		body.Password,
		body.Host,
		body.Port,
		body.Database,
	)

	return sql.Open("mysql", dsn)
}

// doMySQLExecute 执行MySQL查询
func (p *MySQLPlugin) DoSQLExecute(body *Body) (qr *QueryResult) {
	startTime := time.Now()
	defer func() {
		if qr != nil && qr.Message != "success" {
			logger.Log1.WithField("cost", time.Since(startTime).String()).Errorf("SQL查询结束")
		} else {
			logger.Log1.WithField("cost", time.Since(startTime).String()).Infof("SQL查询结束")
		}
	}()

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

	logger.Log1.WithField("sql", body.SQL).Info("执行SQL")
	rows, err := db.Query(body.SQL)
	if err != nil {
		logger.Log1.WithField("error", err).Error("执行SQL查询失败")
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

	// 获取列类型
	columnTypes, errCT := rows.ColumnTypes()
	if errCT != nil {
		logger.Log1.Warningf("获取列类型失败: %v", err)
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
			switch v := (*val).(type) {
			// 对于[]byte类型的数据，特殊处理
			case []byte:
				if p.ValueAsBytes {
					row[col] = v
					continue
				}
				if columnTypes[i].DatabaseTypeName() == "DATE" {
					row[col] = string(v)
				} else {
					row[col] = string(v)
				}
			default:
				row[col] = v
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

func (p *MySQLPlugin) findConfigByKey(key string) *Body {
	for _, config := range p.Configs {
		if config.ConfigKey == key {
			logger.Log1.WithField("config", config).Info("找到配置")
			return &config
		}
	}
	return nil
}

func NewMySQLPlugin() *MySQLPlugin {
	return &MySQLPlugin{
		Name: "mysql_plugin",
	}
}

func (p *MySQLPlugin) Init() error {
	// 定义一个变量来存储 MySQL 配置
	var mysqlConfigs []Body

	// 解析 MySQL 配置
	if err := viper.UnmarshalKey("plugins.mysql", &mysqlConfigs); err != nil {
		logger.Log1.Fatalf("解析 MySQL 配置出错: %v", err)
	}

	p.Configs = mysqlConfigs

	p.AllowRemote = viper.GetBool("auth.mysql.allow_remote")

	logger.Log1.
		WithField("插件名", p.Name).
		WithField("配置列表", p.Configs).
		WithField("允许远程配置", p.AllowRemote).
		WithField("以二进制作为结果", p.ValueAsBytes).
		Info("插件已初始化")
	return nil
}

func (p *MySQLPlugin) HandleMessage(ctx context.Context, df *v1.DFWrap) (*payload.DataFrameResponse, error) {
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

func (p *MySQLPlugin) Close() error {
	// 关闭插件
	logger.Log1.WithField("plugin", p.Name).Info("插件已关闭")
	return nil
}
