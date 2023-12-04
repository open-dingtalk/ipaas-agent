package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"

	"github.com/judwhite/go-svc"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	StreamClientLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-agent/config"
	"go.uber.org/zap"

	_ "github.com/open-dingtalk/ipaas-agent/logger" // 更正后的导入路径
)

func printWelcomePage() {
	// 打印欢迎页面
	fmt.Println("====================================")
	fmt.Println("= Welcome to use ipaas-agent       =")
	fmt.Println("= Version: " + Version + "         =")
	fmt.Println("====================================")
}

type program struct {
	cli *client.StreamClient
}

func (p *program) initClient(env svc.Environment) error {
	clientId := config.GetConfig().Cleint.ClientId
	clientSecret := config.GetConfig().Cleint.ClientSecret

	StreamClientLogger.SetLogger(StreamClientLogger.NewStdTestLogger())

	extra := make(map[string]string)
	// 获取操作系统版本
	extra["osVersion"] = runtime.GOOS
	// 获取主机名
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	extra["hostname"] = hostname

	var openApiHost string
	if runtime.GOOS == "windows" {
		openApiHost = "https://windows-api.dingtalk.com"
	} else {
		openApiHost = "https://pre-api.dingtalk.com"
	}

	p.cli = client.NewStreamClient(
		client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)),
		client.WithOpenApiHost(openApiHost),
		client.WithExtras(extra))

	//注册事件类型的处理函数
	p.cli.RegisterCallbackRouter("/v1.0/ipaas/proxy/callback", func(c context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
		logger := zap.L()
		logger.Info("receive data frame: ", zap.Any("data frame", df.Data))
		response, _ := HandleIpaasCallBack(c, df)
		return response, nil
	})

	return nil
}

func main() {
	prg := &program{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		log.Fatal(err)
	}
}

func (p *program) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		// windows service specific actions
	}

	return p.initClient(env)
}

func (p *program) Start() error {
	printWelcomePage()

	err := p.cli.Start(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (p *program) Stop() error {
	p.cli.Close()
	return nil
}
