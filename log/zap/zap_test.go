package zap_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	z "go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.cloudkitchens.org/lib/log"
	"go.cloudkitchens.org/lib/log/zap"
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
		assert.Equal(t, zapcore.DebugLevel, entry.Level)
		assert.Equal(t, "foo", entry.Message)

		log.Infof(ctx, "foo%v", 2)

		entry = <-entries
		assert.Equal(t, zapcore.InfoLevel, entry.Level)
		assert.Equal(t, "foo2", entry.Message)

		log.Warnf(ctx, "foo%v", 3)

		entry = <-entries
		assert.Equal(t, zapcore.WarnLevel, entry.Level)
		assert.Equal(t, "foo3", entry.Message)

		log.Errorln(ctx, "foo", "4")

		entry = <-entries
		assert.Equal(t, zapcore.ErrorLevel, entry.Level)
		assert.Equal(t, "foo 4\n", entry.Message)
	})

	t.Run("source", func(t *testing.T) {
		call()

		entry := <-entries
		assert.Equal(t, "bar", entry.Message)
		assert.True(t, strings.HasSuffix(entry.Caller.FullPath(), "log/zap/zap_test.go:17"))
	})

}

func ExampleWithContextFields() {
	log.SetLogger(zap.New(z.NewExample(), zap.WithContextFields))
	log.Info(log.NewContext(context.Background(), log.String("foo", "val")), "bar")
	// Output: {"level":"info","msg":"bar","foo":"val"}
}
