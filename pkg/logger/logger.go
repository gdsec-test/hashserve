// Package logger provides contextual logging support via zap.
package logger

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
)

// New will create a new default logger.
func New(paths ...string) (l *zap.Logger, undo func(), err error) {
	prod := zap.NewDevelopmentEncoderConfig()
	prod.NameKey = "name"

	w, wClose, err := zap.Open(paths...)
	if err != nil {
		return nil, nil, err
	}
	l = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(prod),
			w,
			defaultLevel,
		),
	)

	undoLevel := defaultLevel.Level()
	undoGlobal := zap.ReplaceGlobals(l)
	undoStdLog := zap.RedirectStdLog(l)
	undo = func() {
		defer wClose()

		undoGlobal()
		undoStdLog()
		defaultLevel.SetLevel(undoLevel)
	}
	return
}

// Level returns a shared *zap.AtomicLevel to allow changing the level of all
// loggers at runtime.
func Level() zap.AtomicLevel { return defaultLevel }

// SetLevel sets the shared defaultLevel and returns a func to restore it.
func SetLevel(l zapcore.Level) (undo func()) {
	prev := defaultLevel.Level()
	return func() {
		defaultLevel.SetLevel(prev)
	}
}

type contextKey struct{}

// FromContext returns the current zap logger from the given context, or
// the default global logger (by calling zap.L()) otherwise.
func FromContext(ctx context.Context) *zap.Logger {
	if v, ok := ctx.Value(contextKey{}).(*zap.Logger); ok {
		return v
	}
	return zap.L()
}

// WithContext returns a context.Context with a Value containing the given
// zap logger.
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// Check calls zap.Check on the logger for ctx.
func Check(ctx context.Context, lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return FromContext(ctx).Check(lvl, msg)
}

// Debug calls zap.Debug on the logger for ctx.
func Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// Error calls zap.Error on the logger for ctx.
func Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Error(msg, fields...)
}

// Info calls zap.Info on the logger for ctx.
func Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Info(msg, fields...)
}

// Named calls zap.Named on the logger for ctx.
func Named(ctx context.Context, s string) context.Context {
	return WithContext(ctx, FromContext(ctx).Named(s))
}

// Sync calls zap.Sync on the logger for ctx.
func Sync(ctx context.Context) error { return FromContext(ctx).Sync() }

// Warn calls zap.Warn on the logger for ctx.
func Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// With calls zap.With on the logger for ctx.
func With(ctx context.Context, fields ...zapcore.Field) context.Context {
	return WithContext(ctx, FromContext(ctx).With(fields...))
}

// WithOptions calls zap.WithOptions on the logger for ctx.
func WithOptions(ctx context.Context, opts ...zap.Option) context.Context {
	return WithContext(ctx, FromContext(ctx).WithOptions(opts...))
}
