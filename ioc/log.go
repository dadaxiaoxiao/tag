package ioc

import (
	"github.com/dadaxiaoxiao/go-pkg/accesslog"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() accesslog.Logger {
	type Config struct {
		Filename   string `yaml:"filename"`
		Maxsize    int    `yaml:"maxsize"`
		MaxBackups int    `yaml:"maxBackups"`
		MaxAge     int    `yaml:"maxAge"`
	}
	var config Config
	err := viper.UnmarshalKey("logger", &config)
	if err != nil {
		panic(err)
	}

	lumberLogger := &lumberjack.Logger{
		// 要注意，得有权限
		Filename:   config.Filename,
		MaxSize:    config.Maxsize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
		// 同步文档写入
		zapcore.AddSync(lumberLogger),
		zapcore.DebugLevel,
	)
	log := zap.New(core)
	return accesslog.NewZapLogger(log)
}
