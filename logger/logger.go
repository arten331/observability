package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/arten331/observability/logger/oculus_enc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	KeyLevelError = "ERROR"
	KeyLevelInfo  = "INFO"
	KeyLevelDebug = "DEBUG"

	EncodingJSON    = "json"
	EncodingConsole = "console"
	EncodingOculus  = "oculus"

	TimeKey    = "time"
	MessageKey = "message"
	LevelKey   = "level"
	CallerKey  = "caller"

	ConsoleTimeFormat = "2006-01-02T15:04:05.000"
	OculusTimeFormat  = "2006-01-02 15:04:05.999999-07:00"
)

func init() {
	// default unregistered
	MustSetupGlobal(
		WithConfiguration(CoreOptions{
			OutputPath: "stderr",
			Level:      KeyLevelDebug,
			Encoding:   EncodingConsole,
		}),
	)
}

type Logger struct {
	*zap.Logger
	SugaredLogger *zap.SugaredLogger
	CtxLogger     *zap.Logger
	ErrorLogger   *zap.Logger
}

type RotateOptions struct {
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int // days
}

type CoreOptions struct {
	OutputPath string
	Level      string
	Encoding   string
	TimeFormat string
	Rotate     *RotateOptions
}

type Configuration func(l *Logger) error

type Options struct {
	Level string
	Debug bool
}

// _dLogger defaultLogger
var _dLogger Logger

var (
	ErrWrongLogLevelConfiguration = func(opts ...interface{}) error {
		return fmt.Errorf("wrong logger level: %s, instead [%s]", opts...)
	}
	ErrWrongLogEncodingConfiguration = func(opts ...interface{}) error {
		return fmt.Errorf("wrong logger encoding: %s, instead [%s]", opts...)
	}
)

func New(cfgs ...Configuration) (Logger, error) {
	l := Logger{}

	for _, cfg := range cfgs {
		err := cfg(&l)
		if err != nil {
			return l, err
		}
	}

	return l, nil
}

func MustSetupGlobal(cfgs ...Configuration) Logger {
	l, err := New(cfgs...)
	if err != nil {
		log.Fatalf("Unable create global logger, error: %s", err.Error())
	}

	_dLogger = l

	return l
}

// RegisterGlobal replaces the package-global logger after application wiring.
func RegisterGlobal(l *Logger) {
	_dLogger = *l
}

func WithConfiguration(o CoreOptions) Configuration {
	return func(l *Logger) error {
		var (
			err           error
			encoder       zapcore.Encoder
			encoderConfig zapcore.EncoderConfig
			wr            zapcore.WriteSyncer
		)

		dLevel, err := zapcore.ParseLevel(o.Level)
		if err != nil {
			log.Println(
				ErrWrongLogLevelConfiguration(
					o.Level,
					[]string{KeyLevelError, KeyLevelInfo, KeyLevelDebug},
				),
			)

			return err
		}

		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		if o.TimeFormat != "" {
			encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(o.TimeFormat)
		}

		encoderConfig.LevelKey = LevelKey
		encoderConfig.TimeKey = TimeKey
		encoderConfig.MessageKey = MessageKey
		encoderConfig.CallerKey = CallerKey

		switch o.Encoding {
		case EncodingConsole:
			encoderConfig.EncodeLevel = capitalColorLevelEncoder
			encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
			encoderConfig.EncodeTime = TimeEncoderOfLayout(ConsoleTimeFormat)
			encoderConfig.EncodeCaller = ShortCallerEncoder
			encoderConfig.MessageKey = "\t-\t"

			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		case EncodingJSON:
			encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
			encoderConfig.EncodeDuration = zapcore.MillisDurationEncoder

			encoder = zapcore.NewJSONEncoder(encoderConfig)
		case EncodingOculus:
			encoderConfig.MessageKey = "text"
			encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
			encoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
			encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
			encoderConfig.CallerKey = "file"
			encoderConfig.FunctionKey = "function"
			encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(OculusTimeFormat)

			encoder = oculus_enc.NewOculusEncoder(encoderConfig)
		default:
			return ErrWrongLogEncodingConfiguration(o.Encoding, []string{EncodingJSON, EncodingConsole, EncodingOculus})
		}

		switch o.OutputPath {
		case "stdout":
			wr = zapcore.AddSync(os.Stdout)
		case "stderr":
			wr = zapcore.AddSync(os.Stderr)
		default:
			if o.Rotate != nil {
				wr = zapcore.AddSync(&lumberjack.Logger{
					Filename:   o.OutputPath,
					MaxSize:    o.Rotate.MaxSize,
					MaxBackups: o.Rotate.MaxBackups,
					MaxAge:     o.Rotate.MaxAge,
				})

				break
			}

			wr = zapcore.AddSync(&lumberjack.Logger{
				Filename: o.OutputPath,
			})
		}

		core := zapcore.NewCore(encoder, wr, dLevel)

		if l.Logger != nil {
			core = zapcore.NewTee(
				l.Core(),
				zapcore.NewCore(encoder, wr, dLevel),
			)
		}

		l.Logger = zap.New(core, zap.WithCaller(true))
		setupLoggers(l)

		zap.L()

		return nil
	}
}

func L() *Logger {
	return &_dLogger
}

func S() *zap.SugaredLogger {
	return _dLogger.SugaredLogger
}

func CurrentDefault() *Logger {
	return &_dLogger
}

// Named returns a child logger with name appended to every supported logger view.
// The receiver is not mutated, so it is safe to use for scoped dependency injection.
func (l *Logger) Named(name string) *Logger {
	named := *l
	named.Logger = l.Logger.Named(name)
	named.SugaredLogger = l.SugaredLogger.Named(name)
	named.CtxLogger = l.CtxLogger.Named(name)
	named.ErrorLogger = l.ErrorLogger.Named(name)

	return &named
}

func setupLoggers(l *Logger) {
	l.SugaredLogger = l.Sugar()
	l.ErrorLogger = zap.New(l.Core(), zap.AddCallerSkip(1), zap.WithCaller(true))
	l.CtxLogger = zap.New(l.Core(), zap.AddCallerSkip(1))
}

func WithFields(fields ...zap.Field) *Logger {
	l := _dLogger
	l.Logger = l.With(fields...)
	setupLoggers(&l)

	return &l
}
