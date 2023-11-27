package main

import (
	"context"
	"log"
	"syscall"

	"github.com/judwhite/go-svc"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	StreamClientLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"github.com/open-dingtalk/ipaas-net-gateway/config"
	"go.uber.org/zap"

	_ "github.com/open-dingtalk/ipaas-net-gateway/logger"
)

type program struct {
	cli *client.StreamClient
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

	clientId := config.GetConfig().Cleint.ClientId
	clientSecret := config.GetConfig().Cleint.ClientSecret

	StreamClientLogger.SetLogger(StreamClientLogger.NewStdTestLogger())
	p.cli = client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)))

	//注册事件类型的处理函数
	p.cli.RegisterCallbackRouter("/v1.0/ipaas/proxy/callback", func(c context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
		logger := zap.L()
		logger.Info("receive data frame: ", zap.Any("data frame", df.Data))
		response, _ := HandleIpaasCallBack(c, df)
		return response, nil
	})

	return nil
}

func (p *program) Start() error {
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
