package log

import (
	"time"

	"github.com/OdysseyMomentumExperience/token-service/pkg/sentry"
	"github.com/TheZeroSlave/zapsentry"
	sentry_go "github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	DefaultLogger *zap.SugaredLogger
)

type Logger struct {
	unsugared *zap.Logger
	*zap.SugaredLogger
	skipCallerLogger *zap.SugaredLogger
	level            zapcore.Level
}

func NewLogger(level zapcore.Level, sc *sentry.Config, opts ...zap.Option) (*Logger, error) {

	ec := zap.NewDevelopmentEncoderConfig()
	ec.EncodeLevel = zapcore.CapitalColorLevelEncoder

	l, err := zap.Config{
		DisableStacktrace: true,
		Level:             zap.NewAtomicLevelAt(level),
		Development:       true,
		Encoding:          "console",
		EncoderConfig:     ec,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}.Build(opts...)

	if err != nil {
		return nil, err
	}

	if sc.Enable {
		sentry_client, err := sentry_go.NewClient(sentry_go.ClientOptions{
			Dsn:              sc.Dsn,
			Environment:      sc.Environment,
			Release:          sc.Release,
			Debug:            sc.DebugEnable,
			AttachStacktrace: sc.AttachStacktrace,
			BeforeSend: func(event *sentry_go.Event, hint *sentry_go.EventHint) *sentry_go.Event {
				return event
			},
		})
		if err != nil {
			panic(err)
		}
		cfg := zapsentry.Configuration{
			DisableStacktrace: false,
			FlushTimeout:      2 * time.Second,
		}
		core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromClient(sentry_client))
		if err != nil {
			panic(err)
		}
		l = zapsentry.AttachCoreToLogger(core, l)
	}

	logger := &Logger{
		unsugared:        l,
		skipCallerLogger: l.WithOptions(zap.AddCallerSkip(1)).Sugar(),
		SugaredLogger:    l.Sugar(),
		level:            level,
	}
	DefaultLogger = logger.skipCallerLogger

	return logger, nil
}

func (l *Logger) Named(name string, opts ...zap.Option) *Logger {
	unsugared := l.unsugared.Named(name).WithOptions(opts...)

	return &Logger{
		unsugared:        unsugared,
		skipCallerLogger: unsugared.WithOptions(zap.AddCallerSkip(1)).Sugar(),
		SugaredLogger:    unsugared.Sugar(),
		level:            l.level,
	}
}

func (l *Logger) With(args ...interface{}) *Logger {
	sugared := l.SugaredLogger.With(args...)
	unsugared := sugared.Desugar()
	return &Logger{
		unsugared:        unsugared,
		skipCallerLogger: unsugared.WithOptions(zap.AddCallerSkip(1)).Sugar(),
		SugaredLogger:    sugared,
		level:            l.level,
	}
}

func (l *Logger) Error(err error) {
	l.skipCallerLogger.Errorf("error: %+v\n", err)
}

func Debugf(a string, args ...interface{}) {
	DefaultLogger.Debugf(a, args...)
}

func Infof(a string, args ...interface{}) {
	DefaultLogger.Infof(a, args...)
}

func Errorf(a string, args ...interface{}) {
	DefaultLogger.Errorf(a, args...)
}

func Debug(args ...interface{}) {
	DefaultLogger.Debug(args...)
}

func Info(args ...interface{}) {
	DefaultLogger.Info(args...)
}

func Error(err error) {
	DefaultLogger.Errorf("error: %+v\n", err)
}

func Fatal(args ...interface{}) {
	DefaultLogger.Fatal(args...)
}
func Fatalf(a string, args ...interface{}) {
	DefaultLogger.Fatalf(a, args...)
}
