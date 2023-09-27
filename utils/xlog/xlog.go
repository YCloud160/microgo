package xlog

import (
	"context"
	"fmt"
	"github.com/YCloud160/microgo/config"
	"github.com/YCloud160/microgo/meta"
	"github.com/YCloud160/microgo/utils/header"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var (
	logFd *zap.Logger
	level zap.AtomicLevel
)

func InitXlog(conf *config.Config) {
	service := conf.Service
	if len(service) == 0 {
		service = fmt.Sprintf("%d", rand.Uint32()+1000)
	}
	writers := []zapcore.WriteSyncer{os.Stderr}
	output := zapcore.NewMultiWriteSyncer(writers...)
	if len(conf.LogPath) != 0 {
		logFile := filepath.Join(conf.LogPath, service+".log")
		output = zapcore.AddSync(&lumberjack.Logger{
			Filename: logFile,
			MaxSize:  500, // megabytes
			MaxAge:   5,   // days
		})
	}
	encodeConf := zap.NewProductionEncoderConfig()
	encodeConf.TimeKey = "timestamp"
	encodeConf.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	encoder := zapcore.NewJSONEncoder(encodeConf)
	level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	core := zapcore.NewCore(encoder, output, level)
	logFd = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2), zap.AddStacktrace(zapcore.DPanicLevel))
	logFd = logFd.With(zap.Int("pid", os.Getpid()))
	if len(conf.Service) > 0 {
		logFd = logFd.With(zap.String("service", conf.Service))
	}
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	write(ctx, zapcore.DebugLevel, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	write(ctx, zapcore.InfoLevel, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	write(ctx, zapcore.WarnLevel, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	write(ctx, zapcore.ErrorLevel, msg, fields...)
}

func DPanic(ctx context.Context, msg string, fields ...zap.Field) {
	write(ctx, zapcore.DPanicLevel, msg, fields...)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	write(ctx, zapcore.PanicLevel, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	write(ctx, zapcore.FatalLevel, msg, fields...)
}

//func SetLevel(level string) {
//	lowerLevel := strings.ToLower(level)
//	l, ok := LevelName[lowerLevel]
//	if ok {
//		fd.level = l
//	}
//}

func write(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	fields = withContext(ctx, fields...)
	logFd.Log(level, msg, fields...)
}

func Recover(ctx context.Context) {
	if err := recover(); err != nil {
		DPanic(ctx, "recover panic", zap.Any("error", err))
	}
}

func withContext(ctx context.Context, fields ...zap.Field) []zap.Field {
	data, ok := meta.FromOutContext(ctx)
	if !ok {
		return fields
	}
	traceId := data[header.TraceID]
	if len(traceId) > 0 {
		fields = append(fields, zap.String("traceId", traceId))
	}
	spanId := data[header.SpanID]
	if len(spanId) > 0 {
		fields = append(fields, zap.String("spanId", spanId))
	}
	return fields
}
