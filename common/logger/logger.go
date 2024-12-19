package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"path"
	"runtime"
)

type logger struct {
	_logger *zap.Logger
}

func (l *logger) Debug(ctx context.Context, msg string, kv ...interface{}) {
	l.log(zapcore.DebugLevel, msg, append(kv, traceInfo(ctx)...)...)
}

func (l *logger) Info(ctx context.Context, msg string, kv ...interface{}) {
	l.log(zapcore.InfoLevel, msg, append(kv, traceInfo(ctx)...)...)
}

func (l *logger) Warn(ctx context.Context, msg string, kv ...interface{}) {
	l.log(zapcore.WarnLevel, msg, append(kv, traceInfo(ctx)...)...)
}

func (l *logger) Error(ctx context.Context, msg string, kv ...interface{}) {
	l.log(zapcore.ErrorLevel, msg, append(kv, traceInfo(ctx)...)...)
}

func traceInfo(ctx context.Context) []interface{} {
	// 日志行信息中增加追踪参数
	list := make([]interface{}, 0, 6)
	list = append(list, "traceid", ctx.Value("traceid").(string), "spanid", ctx.Value("spanid").(string),
		"pspanid", ctx.Value("pspanid").(string))
	return list
}

// kv 应该是成对的数据, 类似: name,张三,age,10,...
func (l *logger) log(lvl zapcore.Level, msg string, kv ...interface{}) {
	// 保证要打印的日志信息成对出现
	if len(kv)%2 != 0 {
		kv = append(kv, "unknown")
	}

	// 增加日志调用者信息, 方便查日志时定位程序位置
	funcName, file, line := l.getLoggerCallerInfo()
	kv = append(kv, "func", funcName, "file", file, "line", line)

	fields := make([]zap.Field, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		k := fmt.Sprintf("%v", kv[i])
		fields = append(fields, zap.Any(k, kv[i+1]))
	}
	//调用zap.check判断这个日志级别能否写入
	if ce := l._logger.Check(lvl, msg); ce != nil {
		ce.Write(fields...)
	}
}

// getLoggerCallerInfo 日志调用者信息 -- 方法名, 文件名, 行号
func (l *logger) getLoggerCallerInfo() (funcName, file string, line int) {

	pc, file, line, ok := runtime.Caller(3) // 回溯拿调用日志方法的业务函数的信息
	if !ok {
		return
	}
	file = path.Base(file)
	funcName = runtime.FuncForPC(pc).Name()
	return
}

func New() *logger {
	return &logger{
		_logger: _logger,
	}
}
