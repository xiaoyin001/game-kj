// log.go - 日志系统模块
//
// 本文件实现了游戏服务器的日志系统，提供了统一的日志接口和实现
// 使用zap库实现高性能、结构化的日志记录
// 支持多种日志输出目标，包括控制台、文件等
// 支持日志级别控制和日志轮转
// 支持按照日期和小时自动分割日志文件

package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Options 日志配置选项
type Options struct {
	// 日志级别
	Level string
	// 日志基础目录路径
	LogDir string
	// 是否输出到控制台
	Console bool
	// 是否启用开发模式
	Development bool
	// 日志文件最大大小(MB)
	MaxSize int
	// 保留旧日志文件的最大个数
	MaxBackups int
	// 保留旧日志文件的最大天数
	MaxAge int
	// 是否压缩旧日志文件
	Compress bool
}

// 全局单例日志实例
var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	once   sync.Once
	// 用于管理日志文件轮转
	lumberjackWriter *lumberjack.Logger
	// 原子级别控制
	atomicLevel zap.AtomicLevel
)

// InitLogger 初始化全局日志实例
// 这个函数应该在main函数入口处调用
func InitLogger(opts Options) error {
	var err error
	once.Do(func() {
		logger, sugar, err = newLogger(opts)
	})
	return err
}

// newLogger 创建新的日志记录器
func newLogger(opts Options) (*zap.Logger, *zap.SugaredLogger, error) {
	// 设置默认值
	if opts.MaxSize == 0 {
		opts.MaxSize = 100
	}
	if opts.MaxBackups == 0 {
		opts.MaxBackups = 72
	}
	if opts.MaxAge == 0 {
		opts.MaxAge = 28
	}

	// 确保日志目录存在
	if opts.LogDir != "" {
		if err := os.MkdirAll(opts.LogDir, 0755); err != nil {
			return nil, nil, err
		}
	}

	// 解析日志级别
	var level zapcore.Level
	switch opts.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	atomicLevel = zap.NewAtomicLevelAt(level)

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        zapcore.OmitKey,
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 开发模式下使用自定义格式
	if opts.Development {
		// 自定义日志等级编码器 - 缩短格式
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		// 自定义时间格式
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		}

		// 自定义调用者位置格式 - 保持完整路径但减少间距
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		// 使用紧凑型分隔符
		encoderConfig.ConsoleSeparator = " "
	}

	// 创建核心
	cores := []zapcore.Core{}

	// 添加控制台输出
	if opts.Console {
		// 控制台使用自定义配置的编码器
		var consoleEncoder zapcore.Encoder

		if opts.Development {
			// 开发模式下使用自定义紧凑型编码器
			consoleEncoder = newCompactConsoleEncoder(encoderConfig)
		} else {
			// 生产模式使用标准编码器
			consoleEncoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.Lock(os.Stdout),
			atomicLevel,
		)
		cores = append(cores, consoleCore)
	}

	// 添加文件输出
	if opts.LogDir != "" {
		// 获取当前时间
		now := time.Now()
		day := now.Format("2006-01-02")
		hour := now.Hour()

		// 按日期创建子目录
		dayDir := filepath.Join(opts.LogDir, day)
		if err := os.MkdirAll(dayDir, 0755); err != nil {
			return nil, nil, err
		}

		// 按小时创建日志文件
		hourLogFile := filepath.Join(dayDir, fmt.Sprintf("log_%02d.log", hour))

		// 配置日志轮转
		lumberjackWriter = &lumberjack.Logger{
			Filename:   hourLogFile,
			MaxSize:    opts.MaxSize, // MB
			MaxBackups: opts.MaxBackups,
			MaxAge:     opts.MaxAge, // 天
			Compress:   opts.Compress,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(
			fileEncoder,
			zapcore.AddSync(lumberjackWriter),
			atomicLevel,
		)
		cores = append(cores, fileCore)

		// 设置日志文件名更新器（每小时检查一次）
		go updateLogFilename(opts.LogDir)
	}

	// 合并所有核心
	core := zapcore.NewTee(cores...)

	// 创建日志记录器
	zapOptions := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1), // 跳过本包装层
	}

	if opts.Development {
		zapOptions = append(zapOptions, zap.Development())
	}

	// 创建zap logger
	zapLogger := zap.New(core, zapOptions...)
	// 创建sugar logger（提供更友好的API）
	sugarLogger := zapLogger.Sugar()

	return zapLogger, sugarLogger, nil
}

// updateLogFilename 定时更新日志文件名
// 这个函数每小时检查一次是否需要更新日志文件名
func updateLogFilename(logDir string) {
	if lumberjackWriter == nil {
		return
	}

	// 计算下一个整点的时间
	now := time.Now()
	next := now.Add(time.Hour).Truncate(time.Hour)
	duration := next.Sub(now)

	// 先睡眠到下一个整点
	time.Sleep(duration)

	// 然后每小时检查一次
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		day := now.Format("2006-01-02")
		hour := now.Hour()

		// 按日期创建子目录
		dayDir := filepath.Join(logDir, day)
		err := os.MkdirAll(dayDir, 0755)
		if err != nil {
			Error("Failed to create log directory", zap.Error(err))
			continue
		}

		// 按小时创建日志文件
		hourLogFile := filepath.Join(dayDir, fmt.Sprintf("log_%02d.log", hour))

		// 使用lumberjack的Rotate方法关闭当前文件并打开新文件
		lumberjackWriter.Filename = hourLogFile
		err = lumberjackWriter.Rotate()
		if err != nil {
			Error("Failed to rotate log file", zap.Error(err))
		}
	}
}

// Field 类型别名，简化导入
type Field = zap.Field

// 常用Field构造函数
var (
	String   = zap.String
	Int      = zap.Int
	Int64    = zap.Int64
	Float64  = zap.Float64
	Bool     = zap.Bool
	ErrorF   = zap.Error // 改名以避免与全局函数冲突
	Any      = zap.Any
	Duration = zap.Duration
)

// Debug 记录调试级别日志
func Debug(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Debug(msg, fields...)
}

// Info 记录信息级别日志
func Info(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Info(msg, fields...)
}

// Warn 记录警告级别日志
func Warn(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Warn(msg, fields...)
}

// Error 记录错误级别日志
func Error(msg string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Error(msg, fields...)
}

// Fatal 记录致命错误日志并结束程序
func Fatal(msg string, fields ...Field) {
	if logger == nil {
		os.Exit(1)
		return
	}
	logger.Fatal(msg, fields...)
}

// 以下是SugaredLogger的便捷方法，提供更灵活的格式化选项

// Debugf 使用格式化字符串记录调试级别日志
func Debugf(format string, args ...interface{}) {
	if sugar == nil {
		return
	}
	sugar.Debugf(format, args...)
}

// Infof 使用格式化字符串记录信息级别日志
func Infof(format string, args ...interface{}) {
	if sugar == nil {
		return
	}
	sugar.Infof(format, args...)
}

// Warnf 使用格式化字符串记录警告级别日志
func Warnf(format string, args ...interface{}) {
	if sugar == nil {
		return
	}
	sugar.Warnf(format, args...)
}

// Errorf 使用格式化字符串记录错误级别日志
func Errorf(format string, args ...interface{}) {
	if sugar == nil {
		return
	}
	sugar.Errorf(format, args...)
}

// Fatalf 使用格式化字符串记录致命错误日志并结束程序
func Fatalf(format string, args ...interface{}) {
	if sugar == nil {
		os.Exit(1)
		return
	}
	sugar.Fatalf(format, args...)
}

// GetLevel 获取当前日志级别
func GetLevel() zapcore.Level {
	return atomicLevel.Level()
}

// SetLevel 动态设置日志级别
func SetLevel(level zapcore.Level) {
	atomicLevel.SetLevel(level)
}

// Close 关闭日志系统（程序退出前调用）
func Close() error {
	if logger == nil {
		return nil
	}
	return logger.Sync()
}

// 自定义紧凑型控制台编码器工厂函数
func newCompactConsoleEncoder(encoderConfig zapcore.EncoderConfig) zapcore.Encoder {
	// 进一步调整编码器配置，使日志格式更紧凑

	// 确保使用彩色级别显示
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	// 设置精确的时间格式
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	// 自定义调用者路径格式，进一步简化显示
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		// 提取文件名和行号，去掉冗长的路径
		// 示例：从 module/submodule/file.go:123 提取为 file.go:123

		// 提取文件名和行号
		filename := filepath.Base(caller.File)
		lineStr := fmt.Sprintf("%s:%d", filename, caller.Line)
		enc.AppendString(lineStr)
	}

	// 使用超紧凑分隔符
	encoderConfig.ConsoleSeparator = " "

	// 创建一个新的控制台编码器
	return zapcore.NewConsoleEncoder(encoderConfig)
}
