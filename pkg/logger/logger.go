package logger

import (
	"os"
	"sync"

	"github.com/open-dingtalk/ipaas-agent/pkg/ui"
	"github.com/sirupsen/logrus"
)

var (
	Log1 = logrus.New()
	Log2 = logrus.New()
	once sync.Once
)

func init() {
	InitLogger()
}

func InitLogger() {
	// 设置第一个 Logger 输出到文件1
	Log1.SetReportCaller(true)
	file1, err := os.OpenFile("log1.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Failed to open log file 1: %v", err)
	}
	Log1.Out = file1
	Log1.SetLevel(logrus.InfoLevel)
	Log1.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// 设置第二个 Logger 输出到文件2
	Log2.SetReportCaller(true)
	file2, err := os.OpenFile("log2.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Failed to open log file 2: %v", err)
	}
	Log2.Out = file2
	Log2.SetLevel(logrus.InfoLevel)
	Log2.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// 使用 sync.Once 确保 ptermHook 只添加一次
	once.Do(func() {
		Log1.AddHook(&ui.PtermHook{})
	})
}
