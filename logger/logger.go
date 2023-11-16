package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

type LogConfig struct {
	Log struct {
		Level      string `yaml:"level"`
		Path       string `yaml:"path"`
		Name       string `yaml:"name"`
		MaxSize    int    `yaml:"maxsize"`
		MaxAge     int    `yaml:"maxage"`
		MaxBackups int    `yaml:"maxbackups"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"log"`
}

type ZapLoggerWrapper struct {
	logger *zap.SugaredLogger
}

func (z *ZapLoggerWrapper) Write(p []byte) (n int, err error) {
	z.logger.Error(string(p))
	return len(p), nil
}

func init() {
	// Get logger
	logger, err := getLogger("./config/logger.yaml")
	if err != nil {
		panic(err)
	}

	// Replace the global logger
	zap.ReplaceGlobals(logger)
}

func getLogger(configPath string) (*zap.Logger, error) {
	var config LogConfig

	// Read config from file
	data, err := os.ReadFile(configPath)
	if err == nil {
		// Unmarshal config
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, err
		}
	} else {
		// Set default config
		config.Log.Level = "debug"
		config.Log.Path = "./logs"
		config.Log.Name = "gateway.log"
		config.Log.MaxSize = 100
		config.Log.MaxAge = 30
		config.Log.MaxBackups = 10
		config.Log.Compress = true
	}

	// Create a lumberjack logger
	lumberjackLogger := &lumberjack.Logger{
		Filename:   config.Log.Path + "/" + config.Log.Name,
		MaxSize:    config.Log.MaxSize, // megabytes
		MaxBackups: config.Log.MaxBackups,
		MaxAge:     config.Log.MaxAge, //days
		Compress:   config.Log.Compress,
	}

	// Create an encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create a zapcore.Core that writes to our lumberjack logger
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(lumberjackLogger),
		zapcore.DebugLevel,
	)

	// Create a zapcore.Core that writes to the console
	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	// Join the cores together
	core := zapcore.NewTee(fileCore, consoleCore)

	// Create a zap.Logger from the Core
	logger := zap.New(core, zap.AddCaller())

	// Set the logger level
	switch config.Log.Level {
	case "debug":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.DebugLevel))
	case "info":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	case "warn":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.WarnLevel))
	case "error":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	case "dpanic":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.DPanicLevel))
	case "panic":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.PanicLevel))
	case "fatal":
		logger = logger.WithOptions(zap.IncreaseLevel(zap.FatalLevel))
	default:
		logger = logger.WithOptions(zap.IncreaseLevel(zap.DebugLevel))
	}

	return logger, nil
}
