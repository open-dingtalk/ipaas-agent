package main

import (
	"context"

	"github.com/judwhite/go-svc"
	StreamClientLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/ipaas-agent/pkg/client"
	"github.com/open-dingtalk/ipaas-agent/pkg/config"
	"github.com/open-dingtalk/ipaas-agent/pkg/plugins"
	"github.com/open-dingtalk/ipaas-agent/pkg/ui"
	"github.com/sirupsen/logrus"

	"github.com/open-dingtalk/ipaas-agent/pkg/logger"
)

type program struct {
	cli    *client.Client
	ctx    context.Context
	cancel context.CancelFunc
}

var (
	BuildTime string = "DEV"
	GitCommit string = "DEV"
	Version   string = "DEV"
)

func (p *program) Init(env svc.Environment) error {
	// 初始化日志
	// logger.InitLogger()
	// 打印版本信息
	logger.Log1.WithFields(logrus.Fields{
		"buildTime": BuildTime,
		"gitCommit": GitCommit,
		"version":   Version,
	}).Info("启动程序")

	StreamClientLogger.SetLogger(logger.Log2)

	// 读取配置文件
	err := config.LoadConfig()
	if err != nil {
		logger.Log1.Error("加载配置文件出错: ", err)
		return err
	}

	// 初始化UI
	ui.InitUI()

	// 初始化插件
	pluginManager := plugins.NewPluginManager()
	err = pluginManager.LoadPlugins()
	if err != nil {
		logger.Log1.Fatalf("加载插件失败: %v", err)
		return err
	}

	// 初始化客户端
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.cli = client.NewClient(config.GetAuthClientConfig(), pluginManager)
	err = p.cli.Connect(p.ctx)
	connected := err == nil
	if !connected {
		// 这里不退出进程：网络抖动、代理未就绪、凭证下发延迟都会让第一次建连失败，
		// 而 Connect 已经在后台按退避继续重试，退出反而要人工重启才能恢复。
		logger.Log1.WithField("error", err).Error("首次连接服务器失败，网关保持运行并在后台重试")
	}

	// 监听配置文件变化
	go config.WatchConfig(func() {
		logger.Log1.Info("配置文件已更新")
		// 重新加载配置
		pluginManager.ReloadConfig()
	})

	if connected {
		ui.UpdateUISuccess("初始化成功")
	} else {
		ui.UpdateUIError("初始化完成，与服务器的连接将在后台重试")
	}

	return nil
}

func (p *program) Start() error {
	// 启动服务
	go func() {
		<-p.ctx.Done()
		// 在这里处理上下文取消后的清理工作
	}()
	return nil
}

func (p *program) Stop() error {
	// 停止服务
	if p.cli != nil {
		p.cli.Disconnect()
	}
	// 取消上下文
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func main() {
	prg := &program{}
	if err := svc.Run(prg); err != nil {
		logger.Log1.Errorf("服务运行出错: %v", err)
	}
}
