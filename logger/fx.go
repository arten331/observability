package logger

import "go.uber.org/fx"

type moduleParams struct {
	fx.In

	Level    string `name:"log_level"`
	Output   string `name:"log_output"`
	Encoding string `name:"log_encoding"`
}

// Module configures Logger for applications using Uber Fx.
// The regular New constructor remains available for applications without Fx.
func Module() fx.Option {
	return fx.Module(
		"logger",
		fx.Provide(newModuleLogger),
		fx.Invoke(RegisterGlobal),
	)
}

func newModuleLogger(params moduleParams) (*Logger, error) {
	log, err := New(WithConfiguration(CoreOptions{
		Level:      params.Level,
		OutputPath: params.Output,
		Encoding:   params.Encoding,
	}))
	if err != nil {
		return nil, err
	}

	return &log, nil
}
