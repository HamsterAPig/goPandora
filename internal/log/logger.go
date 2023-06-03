package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var (
	log *zap.Logger
)

// InitLogger 初始化日志记录器
func InitLogger(logLevel string) {
	// 自定义时间格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	//自定义日志级别显示
	customLevelEncoder := func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(level.CapitalString())
	}
	// 自定义代码路径、行号输出
	customCallerEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}

	zapLoggerEncoderConfig := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "message",
		StacktraceKey:    "stacktrace",
		EncodeCaller:     customCallerEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeLevel:      customLevelEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		LineEnding:       "\n",
		ConsoleSeparator: " ",
	}

	zapLoggerEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	syncWriter := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))

	// 同步输出到文件
	//syncWriter = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&lumberjack.Logger{
	//	Filename:  "logs/app/app.log", // ⽇志⽂件路径
	//	MaxSize:   100,                // 单位为MB,默认为512MB
	//	MaxAge:    5,                  // 文件最多保存多少天
	//	LocalTime: true,               // 采用本地时间
	//	Compress:  false,              // 是否压缩日志
	//}))
	// 异步输出到文件
	//syncWriter = &zapcore.BufferedWriteSyncer{
	//	WS: zapcore.AddSync(&lumberjack.Logger{
	//		Filename:  "logs/app/app.log", // ⽇志⽂件路径
	//		MaxSize:   100,                // 单位为MB,默认为512MB
	//		MaxAge:    5,                  // 文件最多保存多少天
	//		LocalTime: true,               // 采用本地时间
	//		Compress:  false,              // 是否压缩日志
	//	}),
	//	Size: 4096,
	//}
	zapCore := zapcore.NewCore(zapcore.NewConsoleEncoder(zapLoggerEncoderConfig), syncWriter, parseLogLevel(logLevel))
	log = zap.New(zapCore, zap.AddCaller(), zap.AddCallerSkip(1))
}

// parseLogLevel converts a log level string to the corresponding zapcore.Level.
// It returns -1 if the log level is invalid.
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.Level(-1)
	}
}

// Debug logs a debug level message.
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Info logs an info level message.
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Warn logs a warning level message.
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Error logs an error level message.
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Fatal logs a fatal level message and then exits the program.
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
	os.Exit(1)
}
