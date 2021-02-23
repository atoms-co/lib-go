package zap_test

import (
	"context"
	"strings"
	"testing"

	"go.cloudkitchens.org/lib/log"
	"go.cloudkitchens.org/lib/log/zap"
	"github.com/stretchr/testify/assert"
	z "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func call() {
	log.Info(context.Background(), "bar") // update "source" test if this lines moves
}

func TestLogger(t *testing.T) {
	entries := make(chan zapcore.Entry, 1)

	zl := z.New(zapcore.RegisterHooks(z.NewExample().Core(), func(entry zapcore.Entry) error {
		entries <- entry
		return nil
	}), z.AddCaller())

	log.SetLogger(zap.New(zl))

	t.Run("basic", func(t *testing.T) {
		ctx := context.Background()

		log.Debug(ctx, "foo")

		entry := <-entries
		assert.Equal(t, entry.Level, zapcore.DebugLevel)
		assert.Equal(t, entry.Message, "foo")

		log.Infof(ctx, "foo%v", 2)

		entry = <-entries
		assert.Equal(t, entry.Level, zapcore.InfoLevel)
		assert.Equal(t, entry.Message, "foo2")

		log.Warnf(ctx, "foo%v", 3)

		entry = <-entries
		assert.Equal(t, entry.Level, zapcore.WarnLevel)
		assert.Equal(t, entry.Message, "foo3")

		log.Errorln(ctx, "foo", "4")

		entry = <-entries
		assert.Equal(t, entry.Level, zapcore.ErrorLevel)
		assert.Equal(t, entry.Message, "foo 4\n")
	})

	t.Run("source", func(t *testing.T) {
		call()

		entry := <-entries
		assert.Equal(t, entry.Message, "bar")
		assert.True(t, strings.HasSuffix(entry.Caller.FullPath(), "log/zap/zap_test.go:16"))
	})
}
