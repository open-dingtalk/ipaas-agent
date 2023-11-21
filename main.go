package main

import (
	"context"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
	StreamClientLogger "github.com/open-dingtalk/dingtalk-stream-sdk-go/logger"
	"github.com/open-dingtalk/dingtalk-stream-sdk-go/payload"
	"go.uber.org/zap"

	"github.com/open-dingtalk/ipaas-net-gateway/config"
	_ "github.com/open-dingtalk/ipaas-net-gateway/logger"
)

func main() {
	clientId := config.GetConfig().Cleint.ClientId
	clientSecret := config.GetConfig().Cleint.ClientSecret

	StreamClientLogger.SetLogger(StreamClientLogger.NewStdTestLogger())
	cli := client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(clientId, clientSecret)))

	//注册事件类型的处理函数
	cli.RegisterCallbackRouter("/v1.0/ipaas/proxy/callback", func(c context.Context, df *payload.DataFrame) (*payload.DataFrameResponse, error) {
		logger := zap.L()
		logger.Info("receive data frame: ", zap.Any("data frame", df.Data))
		response, _ := HandleIpaasCallBack(c, df)
		return response, nil
	})

	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}

	defer cli.Close()

	select {}
}
