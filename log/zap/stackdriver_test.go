package zap_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go.cloudkitchens.org/lib/log"
	"go.cloudkitchens.org/lib/log/zap"
	z "go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func callLogger(sev log.Severity, msg string) string {
	ctx := context.Background()
	var buf zaptest.Buffer
	logger := zap.New(z.New(zap.NewStackdriverCore(&buf), z.AddCaller()))

	logger.Log(ctx, sev, 0, msg) // update test if this lines moves
	_ = logger.Flush(ctx)

	return buf.String()
}

type logEntry struct {
	Timestamp string `json:"time"`
}

func TestStackdriverCore(t *testing.T) {
	tests := []struct {
		sev      log.Severity
		msg      string
		expected string
	}{
		{log.SevUnspecified, "foo", `{"severity":"DEBUG","time":"<TIMESTAMP>","message":"foo","logging.googleapis.com/sourceLocation":{"file":"log/zap/stackdriver_test.go","line":20}}` + "\n"},
		{log.SevDebug, "foo", `{"severity":"DEBUG","time":"<TIMESTAMP>","message":"foo","logging.googleapis.com/sourceLocation":{"file":"log/zap/stackdriver_test.go","line":20}}` + "\n"},
		{log.SevInfo, "foo", `{"severity":"INFO","time":"<TIMESTAMP>","message":"foo","logging.googleapis.com/sourceLocation":{"file":"log/zap/stackdriver_test.go","line":20}}` + "\n"},
		{log.SevWarn, "foo", `{"severity":"WARNING","time":"<TIMESTAMP>","message":"foo","logging.googleapis.com/sourceLocation":{"file":"log/zap/stackdriver_test.go","line":20}}` + "\n"},
		{log.SevError, "foo", `{"severity":"ERROR","time":"<TIMESTAMP>","message":"foo","logging.googleapis.com/sourceLocation":{"file":"log/zap/stackdriver_test.go","line":20}}` + "\n"},
		{log.SevFatal, "foo", `{"severity":"CRITICAL","time":"<TIMESTAMP>","message":"foo","logging.googleapis.com/sourceLocation":{"file":"log/zap/stackdriver_test.go","line":20}}` + "\n"},
	}

	for _, test := range tests {
		output := callLogger(test.sev, test.msg)

		// We can't control zap time, so we replace the expected TIMESTAMP with the actual zap timestamp.
		var entry logEntry
		_ = json.Unmarshal([]byte(output), &entry)
		expected := strings.ReplaceAll(test.expected, "<TIMESTAMP>", entry.Timestamp)

		if output != expected {
			t.Errorf("Incorrect message:\n(%v), want:\n(%v)", output, expected)
		}
	}
}
