package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 日志实例
var loggerInstance *zap.Logger
var log *zap.SugaredLogger

// ANSI 颜色代码
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[36m"
	colorGray   = "\033[37m"
)

// customCore 实现 zapcore.Core 以提供与 logrus 一模一样的格式
type customCore struct {
	level  zapcore.LevelEnabler
	writer zapcore.WriteSyncer
}

func (c *customCore) Enabled(l zapcore.Level) bool {
	return c.level.Enabled(l)
}

func (c *customCore) With(fields []zapcore.Field) zapcore.Core {
	return c
}

func (c *customCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *customCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	timestamp := ent.Time.Format("2006-01-02 15:04:05")
	level := strings.ToUpper(ent.Level.String())

	var levelColor string
	switch ent.Level {
	case zapcore.DebugLevel:
		levelColor = colorGray
	case zapcore.InfoLevel:
		levelColor = colorBlue
	case zapcore.WarnLevel:
		levelColor = colorYellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		levelColor = colorRed
	default:
		levelColor = colorBlue
	}

	msg := fmt.Sprintf("[%s]%s[%s]%s %s\n", timestamp, levelColor, level, colorReset, ent.Message)
	_, err := c.writer.Write([]byte(msg))
	return err
}

func (c *customCore) Sync() error {
	return c.writer.Sync()
}

func initLogger(logFile string, fileOnly bool) {
	logDir := filepath.Dir(logFile)
	if logDir != "" && logDir != "." {
		os.MkdirAll(logDir, 0755)
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     0,
		Compress:   false,
	}

	var output zapcore.WriteSyncer
	// fileOnly 模式下只输出到文件（daemon 模式或重启模式）
	if fileOnly {
		output = zapcore.AddSync(lumberjackLogger)
	} else {
		// 前台运行时同时输出到终端和文件
		output = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberjackLogger))
	}

	core := &customCore{
		level:  zap.NewAtomicLevelAt(zap.InfoLevel),
		writer: output,
	}

	loggerInstance = zap.New(core)
	log = loggerInstance.Sugar()
}
