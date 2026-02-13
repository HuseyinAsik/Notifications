package logging

import (
	"context"
	"os"
	"time"

	"github.com/HuseyinAsik/Notifications/pkg/settings"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *LogWrapper

func Setup(config *settings.App) {
	var options []zap.Option
	logWriter := zapcore.AddSync(os.Stdout)

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
	atomicLevel := zap.NewAtomicLevel()
	level, exist := loggerLevelMap[config.LogLevel]
	if !exist {
		level = zapcore.DebugLevel
	}
	atomicLevel.SetLevel(level)

	var encoder zapcore.Encoder
	if config.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, logWriter, atomicLevel)
	serviceNameField := zap.Fields(zap.String("serviceName", config.AppName))
	hostNameField := zap.Fields(zap.String("hostname", config.Hostname))

	options = append(options, zap.AddCaller(), serviceNameField, hostNameField, zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	if config.RunMode != "release" {
		options = append(options, zap.Development())
	}
	logger = &LogWrapper{ZapLogger: zap.New(core, options...)}
}

func GetLogger() *LogWrapper {
	return logger
}

var loggerLevelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
}

type LogWrapper struct {
	ZapLogger *zap.Logger
}
type LogModel struct {
	fields []zap.Field
}

func (l *LogWrapper) enrichFields(ctx context.Context, fields []zap.Field) []zap.Field {
	fields = append(fields, zap.Time("time", time.Now().UTC()))

	if lm := ctx.Value("logModel"); lm != nil {
		if logModel, ok := lm.(LogModel); ok {
			logModel.fields = append(logModel.fields, fields...)
			return logModel.fields
		}
	}

	return fields
}

func (l *LogWrapper) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = l.enrichFields(ctx, fields)
	l.ZapLogger.Debug(msg, fields...)
}

func (l *LogWrapper) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = l.enrichFields(ctx, fields)
	l.ZapLogger.Info(msg, fields...)
}

func (l *LogWrapper) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = l.enrichFields(ctx, fields)
	l.ZapLogger.Warn(msg, fields...)
}

func (l *LogWrapper) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = l.enrichFields(ctx, fields)
	l.ZapLogger.Error(msg, fields...)
}

func (l *LogWrapper) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = l.enrichFields(ctx, fields)
	l.ZapLogger.Fatal(msg, fields...)
}

func (l *LogWrapper) Sync() error {
	return l.ZapLogger.Sync()
}
