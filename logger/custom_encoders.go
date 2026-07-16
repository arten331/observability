package logger

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

var (
	_levelToColor = map[zapcore.Level]Color{
		zapcore.DebugLevel: White,
		zapcore.InfoLevel:  Green,
		zapcore.ErrorLevel: Red,
	}
	_unknownLevelColor = Red

	_levelToCapitalColorString = make(map[zapcore.Level]string, len(_levelToColor))

	_lastMaxAlign = 24
)

// Color represents a text color.
type Color uint8

// Add adds the coloring to the given string.
func (c Color) Add(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}

func init() {
	for level, color := range _levelToColor {
		_levelToCapitalColorString[level] = color.Add(level.CapitalString())
	}
}

func capitalColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := _levelToCapitalColorString[l]
	if !ok {
		s = _unknownLevelColor.Add(l.CapitalString())
	}
	enc.AppendString(s)
}

func TimeEncoderOfLayout(layout string) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		var buffer bytes.Buffer

		buffer.WriteString("[")
		buffer.WriteString(Cyan.Add(t.Format(ConsoleTimeFormat)))
		buffer.WriteString("]")

		enc.AppendString(buffer.String())
	}
}

func ShortCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(Magenta.Add(rightPadCaller(caller.TrimmedPath())))
}

func rightPadCaller(s string) string {
	padStr := " "

	overallLen := _lastMaxAlign

	if len(s) > _lastMaxAlign {
		if len(s)%4 == 0 {
			_lastMaxAlign = len(s)
		} else {
			_lastMaxAlign = len(s) + len(s)%4
		}
	}

	return (s + strings.Repeat(padStr, 1+((_lastMaxAlign-1)/1)))[:overallLen]
}
