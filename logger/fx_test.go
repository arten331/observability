package logger

import (
	"testing"

	"go.uber.org/fx"
)

func TestModule(t *testing.T) {
	var configured *Logger

	app := fx.New(
		fx.NopLogger,
		Module(),
		fx.Provide(
			fx.Annotate(
				func() string { return KeyLevelDebug },
				fx.ResultTags(`name:"log_level"`),
			),
			fx.Annotate(
				func() string { return "stderr" },
				fx.ResultTags(`name:"log_output"`),
			),
			fx.Annotate(
				func() string { return EncodingConsole },
				fx.ResultTags(`name:"log_encoding"`),
			),
		),
		fx.Populate(&configured),
	)

	if err := app.Start(t.Context()); err != nil {
		t.Fatalf("start Fx app: %v", err)
	}

	defer func() {
		if err := app.Stop(t.Context()); err != nil {
			t.Errorf("stop Fx app: %v", err)
		}
	}()

	if configured == nil || configured.Logger == nil {
		t.Fatal("Module() did not provide a configured logger")
	}

	if CurrentDefault().Logger != configured.Logger {
		t.Fatal("Module() did not register the configured logger globally")
	}
}
