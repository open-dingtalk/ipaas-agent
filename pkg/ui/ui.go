package ui

import (
	"sync"

	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
)

var (
	once              sync.Once
	AgentMultiPrinter *pterm.MultiPrinter
	title             string = "IPaaS 本地网关"
)

type PtermHook struct{}

func (hook *PtermHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *PtermHook) Fire(entry *logrus.Entry) error {
	message := entry.Message
	logger := pterm.DefaultLogger.
		WithLevel(pterm.LogLevelTrace)
	switch entry.Level {
	case logrus.InfoLevel:
		logger.Info(message)
	case logrus.WarnLevel:
		logger.Warn(message)
	case logrus.ErrorLevel:
		logger.Error(message, logger.Args("caller", entry.Caller.Function))
	case logrus.FatalLevel:
		logger.Fatal(message, logger.Args("caller", entry.Caller.Function))
	case logrus.DebugLevel:
		logger.Debug(message)
	default:
		logger.Info(message)
	}
	return nil
}

func InitUI() {
	once.Do(func() {
		// 可以在这里进行全局的 UI 初始化配置
		// pterm.EnableDebugMessages() // 启用调试信息（可选）
		AgentMultiPrinter = &pterm.DefaultMultiPrinter
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithFullWidth().Println(title)
		pterm.Println()
	})
}

func UpdateUISuccess(message string) {
	// 使用 pterm 显示信息
	pterm.Success.Println(message)
}
