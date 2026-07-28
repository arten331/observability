package oculus_enc

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

func (enc *oculusEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()
	final.buf.AppendByte('{')

	if final.MessageKey != "" {
		final.addKey(enc.MessageKey)
		final.AppendString(ent.Message)
	}

	if enc.buf.Len() > 0 {
		final.addElementSeparator()
		final.buf.AppendBytes(enc.buf.Bytes())
	}

	final.addKey("record")
	final.buf.AppendByte('{')

	if ent.Caller.Defined {
		if final.CallerKey != "" {
			final.addKey(final.CallerKey)
			final.buf.AppendByte('{')

			final.addKey("name")
			cur := final.buf.Len()
			final.EncodeCaller(ent.Caller, final)
			if cur == final.buf.Len() {
				final.AppendString(ent.Caller.String())
			}

			final.buf.AppendByte('}')
			final.addElementSeparator()
		}

		if final.FunctionKey != "" {
			final.addKey(final.FunctionKey)
			final.AppendString(ent.Caller.Function)
		}
	}

	final.addElementSeparator()

	if final.LevelKey != "" && final.EncodeLevel != nil {
		final.addKey(final.LevelKey)
		final.buf.AppendByte('{')
		final.addKey("name")
		final.EncodeLevel(ent.Level, final)
		final.buf.AppendByte('}')
	}

	final.addElementSeparator()

	if final.TimeKey != "" {
		final.addKey(final.TimeKey)

		final.buf.AppendByte('{')

		final.addKey("repr")
		final.EncodeTime(ent.Time, final)

		final.buf.AppendByte('}')
	}

	final.addElementSeparator()

	final.addKey("extra")

	final.buf.AppendByte('{')
	if len(fields) > 0 {
		addFields(final, fields)
	}

	final.buf.AppendByte('}')

	final.buf.AppendByte('}')
	final.buf.AppendByte('}')
	final.buf.AppendString(final.LineEnding)

	ret := final.buf
	putOculusEncoder(final)
	return ret, nil
}
