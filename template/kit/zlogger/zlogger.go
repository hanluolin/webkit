package zlogger

import (
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Option struct {
	Path       string // 日志文件路径
	Level      string // 日志级别，debug info warn error panic fatal
	MaxSize    int    // 文件多大开始切分
	MaxBackups int    // 保留文件个数
	MaxAge     int    // 文件保存多少天，maxBackups和maxAge都设置为0，则不会删除任何日志文件，全部保留
	Json       bool   // 是否用json格式
	Std        bool   // 是否输出到控制台
}

func DefaultLog() *Option {
	return &Option{
		Path:    "",
		Level:   "debug",
		MaxSize: 10,
		Json:    false,
		Std:     true, // 终端是否输出
	}
}

// Init 配置zap日志、lumberjack日志切割归档，并将设置后的zap日志全局置入 后续的程序中如果需要写日志 直接zap.L() || zap.S()即可
func Init(option *Option) {
	var (
		logger        *zap.Logger
		encoder       zapcore.Encoder
		encoderConfig = zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "logger",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeTime:     timeEncoder, // ISO8601 UTC 时间格式
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
			EncodeName:     zapcore.FullNameEncoder,
		}
	)

	if option.MaxSize == 0 {
		option.MaxSize = 10
	}

	// 设置日志输出格式 (json / console)
	if option.Json {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 添加日志切割归档功能
	var (
		outPut zapcore.WriteSyncer
		hook   = lumberjack.Logger{
			Filename:   option.Path,       // 日志文件路径
			MaxSize:    option.MaxSize,    // 每个日志文件保存的最大尺寸 单位：M
			MaxBackups: option.MaxBackups, // 日志文件最多保存多少个备份
			MaxAge:     option.MaxAge,     // 文件最多保存多少天
			Compress:   true,              // 是否压缩
		}
	)
	// 是否输出到控制台
	if option.Std {
		outPut = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stderr), zapcore.AddSync(&hook))
	} else {
		outPut = zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook))
	}

	var core = zapcore.NewCore(
		encoder, // 编码器配置
		outPut,  // 打印到控制台和文件
		zap.NewAtomicLevelAt(parseLevel(option.Level)), // 日志级别
	)

	logger = zap.New(core, zap.AddCaller())
	// logger = zap.New(core) 不追踪堆栈

	// 将自定义的logger替换为全局的logger
	zap.ReplaceGlobals(logger)
}

func parseLevel(val string) zapcore.Level {
	switch val {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.DebugLevel
	}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func Sync() {
	zap.L().Sync()
	zap.S().Sync()
}
