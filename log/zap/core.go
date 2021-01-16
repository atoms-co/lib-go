package zap

import (
	z "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Transform is an entry and fields transformation function.
type Transform func(zapcore.Entry, []zapcore.Field) (zapcore.Entry, []zapcore.Field)

type transform struct {
	z.AtomicLevel
	fns  []Transform
	core zapcore.Core
}

// NewTransformCore returns a Core wrapper that transforms each entry and fields as specified.
func NewTransformCore(level zapcore.Level, core zapcore.Core, fns ...Transform) zapcore.Core {
	if core == nil {
		panic("nil core")
	}
	if len(fns) == 0 {
		return core
	}
	return &transform{
		AtomicLevel: z.NewAtomicLevelAt(level),
		fns:         fns,
		core:        core,
	}
}

func (c *transform) With(fields []zapcore.Field) zapcore.Core {
	return NewTransformCore(c.AtomicLevel.Level(), c.core.With(fields), c.fns...)
}

func (c *transform) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}
	return checkedEntry
}

func (c *transform) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	for _, fn := range c.fns {
		entry, fields = fn(entry, fields)
	}
	return c.core.Write(entry, fields)
}

func (c *transform) Sync() error {
	return c.core.Sync()
}
