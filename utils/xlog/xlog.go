package xlog

import (
	"context"
	"io"
	"os"
	"runtime"
	"strings"
)

type Logger interface {
	Debug(msg string, fields ...*Entry)
	Info(msg string, fields ...*Entry)
	Warn(msg string, fields ...*Entry)
	Error(msg string, fields ...*Entry)
	Fatal(msg string, fields ...*Entry)
}

type LogOption func(fd *logFD)

type logFD struct {
	level    Level
	writers  []io.Writer
	fields   []*Entry
	ctxWrite func(ctx context.Context) []*Entry
}

var fd *logFD

func init() {
	fd = &logFD{
		writers: []io.Writer{os.Stdout},
		level:   INFO,
		fields:  make([]*Entry, 0),
	}
}

func InitXlog(options ...LogOption) {
	for _, opt := range options {
		opt(fd)
	}
}

func WithWriter(writers ...io.Writer) LogOption {
	return func(fd *logFD) {
		fd.writers = writers
	}
}

func WithLevel(level Level) LogOption {
	return func(fd *logFD) {
		fd.level = level
	}
}

func WithField(fields ...*Entry) LogOption {
	return func(fd *logFD) {
		fd.fields = append(fd.fields, fields...)
	}
}

func WithContextWrite(f func(ctx context.Context) []*Entry) LogOption {
	return func(fd *logFD) {
		fd.ctxWrite = f
	}
}

func Debug(ctx context.Context, msg string, fields ...*Entry) {
	write(ctx, DEBUG, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...*Entry) {
	write(ctx, INFO, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...*Entry) {
	write(ctx, WARN, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...*Entry) {
	write(ctx, ERROR, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...*Entry) {
	write(ctx, FATAL, msg, fields...)
}

func SetLevel(level string) {
	lowerLevel := strings.ToLower(level)
	l, ok := LevelName[lowerLevel]
	if ok {
		fd.level = l
	}
}

func write(ctx context.Context, level Level, msg string, fields ...*Entry) {
	if fd.level > level {
		return
	}
	entry := make([]*Entry, 0, len(fields)+len(fd.fields)+10)
	entry = append(entry, timeField(), levelField(level), caller(), Field("msg", msg))

	if fd.ctxWrite != nil {
		ctxFields := fd.ctxWrite(ctx)
		if len(ctxFields) > 0 {
			entry = append(entry, ctxFields...)
		}
	}
	if len(fd.fields) > 0 {
		entry = append(entry, fd.fields...)
	}
	entry = append(entry, fields...)
	if level == FATAL {
		entry = append(entry, Field("stack", stack()))
	}

	logMessage := EncodeJson(entry...)
	for _, w := range fd.writers {
		w.Write([]byte(logMessage + "\n"))
	}
	if level == FATAL {
		os.Exit(0)
	}
}

func Recover(ctx context.Context) {
	if err := recover(); err != nil {
		Error(ctx, "recover panic", Field("error", err), Field("stack", stack()))
	}
}

func stack() string {
	var buf [2 << 10]byte
	return string(buf[:runtime.Stack(buf[:], false)])
}
