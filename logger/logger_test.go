//go:build test && !integration

package logger

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type MarshalEvent struct {
	Code   string `json:"code"`
	Ip     string `json:"ip_address"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (e MarshalEvent) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("cd", e.Code)
	encoder.AddString("ip", e.Ip)
	encoder.AddString("nm", e.Name)
	encoder.AddString("st", e.Status)

	return nil
}

type ReflectEvent struct {
	Code   string `json:"code"`
	Ip     string `json:"ip_address"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type TestCase struct {
	desc            string
	options         []CoreOptions
	data            []zap.Field
	expectedConsole []string
	expectedJson    []string
}

func TestLogger(t *testing.T) {
	var testFields = []zap.Field{
		zap.String("string", "smileEveryday"),
		zap.Int("integer", 300),
		zap.Object("objMarshaleld", MarshalEvent{
			Code:   "500",
			Ip:     "127.0.0.1",
			Name:   "Created",
			Status: "Error",
		}),
		zap.Reflect("reflect", ReflectEvent{
			Code:   "500",
			Ip:     "127.0.0.1",
			Name:   "Created",
			Status: "Error",
		}),
		zap.Error(errors.New("AS Error")),
	}

	testLogFile := "/tmp/testLogger/" + strconv.FormatInt(time.Now().Unix(), 10) + ".log"
	defer func(name string) {
		_ = os.Remove(name)
	}(testLogFile)

	testCases := []TestCase{
		{
			desc: "Console inline",
			options: []CoreOptions{
				/* TODO: I can't catch a log from stdout/err,  only from a file */
				//{
				//	OutputPath: "stdout",
				//	Level:      KeyLevelDebug,
				//	Encoding:   EncodingConsole,
				//},
				//{
				//	OutputPath: "stderr",
				//	Level:      KeyLevelDebug,
				//	Encoding:   EncodingJSON,
				//},
				{
					OutputPath: testLogFile,
					Level:      KeyLevelDebug,
					Encoding:   EncodingJSON,
					Rotate: &RotateOptions{
						MaxSize:    100,
						MaxBackups: 4,
						MaxAge:     7,
					},
				},
			},
			data:            testFields,
			expectedConsole: []string{`"objMarshaleld": {"cd": "500", "ip": "127.0.0.1", "nm": "Created", "st": "Error"}, "reflect": {"code":"500","ip_address":"127.0.0.1","name":"Created","status":"Error"}`},
			expectedJson:    []string{`"objMarshaleld":{"cd":"500","ip":"127.0.0.1","nm":"Created","st":"Error"},"reflect":{"code":"500","ip_address":"127.0.0.1","name":"Created","status":"Error"`},
		},
	}

	for _, tc := range testCases {

		var lConf []Configuration

		for i := range tc.options {
			lConf = append(lConf, WithConfiguration(tc.options[i]))
		}

		logger, _ := New(lConf...)

		for i := range tc.options {
			var captured string

			switch tc.options[i].OutputPath {
			case "stdout":
				captured = captureStdout(func() {
					logger.Logger.Info("Info message", tc.data...)
				})
			case "stderr":
				captured = captureStdErr(func() {
					logger.Logger.Error("Info message", tc.data...)
					logger.Logger.Debug("Info message", tc.data...)
					logger.Logger.Info("Info message", tc.data...)
				})
			default:
				logger.Logger.Debug("Info message", tc.data...)
				b, err := ioutil.ReadFile(testLogFile)
				if err != nil {
					t.Fatalf("Unable open test log file %s", testLogFile)
				}

				captured = string(b)
			}

			if len(captured) == 0 {
				t.Errorf("Failed to catch the line in: %s", tc.options[i].OutputPath)
			} else {
				t.Logf("Catch line in %s", tc.options[i].OutputPath)
			}

			var expectedStrings []string

			switch tc.options[i].Encoding {
			case EncodingJSON:
				expectedStrings = tc.expectedJson
			case EncodingConsole:
				expectedStrings = tc.expectedConsole
			default:
				t.Fatalf("Wrong encoding type")
			}

			for _, es := range expectedStrings {
				includeExpecting := strings.Contains(captured, es)
				if includeExpecting != true {
					t.Errorf("Expected string not found in: %s", tc.options[i].OutputPath)
				}
			}

		}
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	<-time.After(1000 * time.Millisecond)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func captureStdErr(f func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	f()

	<-time.After(1000 * time.Millisecond)

	_ = w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func BenchmarkMarshaled(b *testing.B) {

	MustSetupGlobal(WithConfiguration(
		CoreOptions{
			OutputPath: "stderr",
			Level:      KeyLevelDebug,
			Encoding:   EncodingJSON,
		},
	))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		L().Info("", zap.Object("objMarshaleld", MarshalEvent{
			Code:   "500",
			Ip:     "127.0.0.1",
			Name:   "Created",
			Status: "Error",
		}))
		_ = L().Sync()
	}
}

func BenchmarkReflected(b *testing.B) {

	MustSetupGlobal(WithConfiguration(
		CoreOptions{
			OutputPath: "stderr",
			Level:      KeyLevelDebug,
			Encoding:   EncodingJSON,
		},
	))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		L().Info("", zap.Reflect("objReflected", ReflectEvent{
			Code:   "500",
			Ip:     "127.0.0.1",
			Name:   "Created",
			Status: "Error",
		}))
		_ = L().Sync()
	}

}

func BenchmarkMarshaledConsole(b *testing.B) {

	MustSetupGlobal(WithConfiguration(
		CoreOptions{
			OutputPath: "stderr",
			Level:      KeyLevelDebug,
			Encoding:   EncodingConsole,
		},
	))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		L().Info("", zap.Object("objMarshaleld", MarshalEvent{
			Code:   "500",
			Ip:     "127.0.0.1",
			Name:   "Created",
			Status: "Error",
		}))
		_ = L().Sync()
	}
}

func BenchmarkReflectedConsole(b *testing.B) {

	MustSetupGlobal(WithConfiguration(
		CoreOptions{
			OutputPath: "stderr",
			Level:      KeyLevelDebug,
			Encoding:   EncodingConsole,
		},
	))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		L().Info("", zap.Reflect("objReflected", ReflectEvent{
			Code:   "500",
			Ip:     "127.0.0.1",
			Name:   "Created",
			Status: "Error",
		}))
		_ = L().Sync()
	}

}
