package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger      *zap.Logger
	GlobalLevel = zap.NewAtomicLevel()
)

func formatEncodeTime(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()))
}

// getLogWriter 配置日志轮转
func getLogWriter(logPath string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100,   // 单个日志文件最大100MB
		MaxBackups: 7,     // 保留7个备份
		MaxAge:     7,     // 保留7天
		Compress:   false, // 不压缩
	}
	return zapcore.AddSync(lumberJackLogger)
}

func InitLogger(logLevel string) {
	// 创建日志目录
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("create log dir failed: " + err.Error())
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     formatEncodeTime,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// logLevel转小写
	logLevel = strings.ToLower(logLevel)
	switch logLevel {
	case "debug":
		GlobalLevel.SetLevel(zap.DebugLevel)
	case "info":
		GlobalLevel.SetLevel(zap.InfoLevel)
	case "warn":
		GlobalLevel.SetLevel(zap.WarnLevel)
	case "error":
		GlobalLevel.SetLevel(zap.ErrorLevel)
	default:
		GlobalLevel.SetLevel(zap.DebugLevel)
	}

	// 多级别日志配置
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	// 创建核心
	core := zapcore.NewTee(
		// INFO级别日志（按大小和天数轮转）
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			getLogWriter(filepath.Join(logDir, "sip.log")),
			GlobalLevel, // 使用全局可调级别
		),
		// ERROR级别日志（单独文件）
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			getLogWriter(filepath.Join(logDir, "error_sip.log")),
			errorLevel,
		),
		// 控制台输出
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		),
	)

	// 创建Logger
	Logger = zap.New(core, zap.AddCaller())
	defer Logger.Sync() // 程序退出前刷新缓冲区
}
