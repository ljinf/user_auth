package logger

import (
	"github.com/ljinf/user_auth/common/enum"
	"github.com/ljinf/user_auth/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

var _logger *zap.Logger

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	logWriter := getFileLogWriter()

	var cores []zapcore.Core
	switch config.App.Env {
	case enum.ModeTest, enum.ModeProd:
		// 测试环境和生产环境的日志输出到文件中
		cores = append(cores, zapcore.NewCore(encoder, logWriter, zap.InfoLevel))
		break
	case enum.ModeDev:
		// 开发环境同时向控制台和文件输出日志， Debug级别的日志也会被输出
		cores = append(cores,
			zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zap.DebugLevel),
			zapcore.NewCore(encoder, logWriter, zap.DebugLevel))
	}
	core := zapcore.NewTee(cores...)
	_logger = zap.New(core)
}

func getFileLogWriter() zapcore.WriteSyncer {
	// 使用 lumberjack 实现 logger rotate
	logger := &lumberjack.Logger{
		Filename:  config.App.Log.FilePath,
		MaxSize:   config.App.Log.FileMaxSize,      //文件大小
		MaxAge:    config.App.Log.BackUpFileMaxAge, //旧文件最大保留天数
		Compress:  false,
		LocalTime: true,
	}

	return zapcore.AddSync(logger)
}
