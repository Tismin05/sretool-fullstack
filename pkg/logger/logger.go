package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局日志实例
var Logger *zap.SugaredLogger

// Config 日志配置
type Config struct {
	Level      string `mapstructure:"level"`       // 日志级别: debug, info, warn, error
	Format     string `mapstructure:"format"`      // 输出格式: json, console
	OutputPath string `mapstructure:"output_path"` // 输出路径
}

// Init 初始化日志
func Init(cfg Config) error {
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置编码器
	var encoder zapcore.Encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.OutputPath != "" {
		file, err := os.OpenFile(cfg.OutputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Logger = zapLogger.Sugar()

	return nil
}

// Debug 调试日志
func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

// Info 信息日志
func Info(args ...interface{}) {
	Logger.Info(args...)
}

// Warn 警告日志
func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

// Error 错误日志
func Error(args ...interface{}) {
	Logger.Error(args...)
}

// Debugf 调试日志（格式化）
func Debugf(template string, args ...interface{}) {
	Logger.Debugf(template, args...)
}

// Infof 信息日志（格式化）
func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args...)
}

// Warnf 警告日志（格式化）
func Warnf(template string, args ...interface{}) {
	Logger.Warnf(template, args...)
}

// Errorf 错误日志（格式化）
func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args...)
}

// Sync 同步日志缓冲
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
