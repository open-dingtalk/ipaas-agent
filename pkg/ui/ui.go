package ui

import (
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
)

var (
	headerArea *pterm.AreaPrinter
	once       sync.Once
	title      string = "IPaaS 本地网关"
)

type PtermHook struct{}

type SQLSpinner struct {
	*pterm.SpinnerPrinter
	StartTime time.Time
}

func InitUI() {
	once.Do(func() {
		// 创建持久显示的区域
		headerArea, _ = pterm.DefaultArea.Start()

		// 更新头部区域
		updateHeader()

		pterm.Println() // 为内容留出空间
	})
}

func updateHeader() {
	var headerContent string

	// 渲染头部
	header := pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).WithFullWidth()
	headerStr := header.Sprint(title)

	headerContent = headerStr

	// 更新区域内容
	headerArea.Update(headerContent)
}

func UpdateUISuccess(message string) {
	// updateHeader() // 重新渲染头部
	pterm.Success.Println(message)
}

func (hook *PtermHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *PtermHook) Fire(entry *logrus.Entry) error {
	message := entry.Message
	fields := entry.Data
	// 将 logrus fields 转换为字符串切片
	var args []interface{}
	for k, v := range fields {
		args = append(args, k, v)
	}
	logger := pterm.DefaultLogger.
		WithLevel(pterm.LogLevelTrace)
	switch entry.Level {
	case logrus.InfoLevel:
		logger.Info(message, logger.Args(args...))
	case logrus.WarnLevel:
		logger.Warn(message, logger.Args(args...))
	case logrus.ErrorLevel:
		logger.Error(message, logger.Args(args...), logger.Args("caller", entry.Caller.Function))
	case logrus.FatalLevel:
		logger.Fatal(message, logger.Args(args...), logger.Args("caller", entry.Caller.Function))
	case logrus.DebugLevel:
		logger.Debug(message, logger.Args(args...))
	default:
		logger.Info(message)
	}
	return nil
}

func StartSpinner(message string) *pterm.SpinnerPrinter {
	pterm.DefaultSpinner.RemoveWhenDone = true
	spinner, _ := pterm.DefaultSpinner.Start(message)
	return spinner
}
