package zap_test

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"go.atoms.co/lib/log"
	"go.atoms.co/lib/log/zap"
	z "go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func callLogger(sev log.Severity, msg string) io.Reader {
	ctx := context.Background()
	var buf zaptest.Buffer
	logger := zap.New(z.New(zap.NewStackdriverCore(&buf), z.AddCaller()))

	logger.Log(ctx, sev, 0, msg) // update test if this lines moves
	_ = logger.Flush(ctx)

	return &buf.Buffer
}

type logEntry struct {
	Severity       string         `json:"severity"`
	Timestamp      string         `json:"time"`
	Message        string         `json:"message"`
	SourceLocation sourceLocation `json:"logging.googleapis.com/sourceLocation"`
}

type sourceLocation struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

func TestStackdriverCore(t *testing.T) {
	tests := []struct {
		sev      log.Severity
		msg      string
		expected logEntry
	}{
		{
			log.SevUnspecified,
			"foo",
			logEntry{
				Severity: "DEBUG",
				Message:  "foo",
				SourceLocation: sourceLocation{
					File: "log/zap/stackdriver_test.go",
					Line: 21,
				},
			},
		},
		{
			log.SevDebug,
			"foo",
			logEntry{
				Severity: "DEBUG",
				Message:  "foo",
				SourceLocation: sourceLocation{
					File: "log/zap/stackdriver_test.go",
					Line: 21,
				},
			},
		},
		{
			log.SevInfo,
			"foo",
			logEntry{
				Severity: "INFO",
				Message:  "foo",
				SourceLocation: sourceLocation{
					File: "log/zap/stackdriver_test.go",
					Line: 21,
				},
			},
		},
		{
			log.SevWarn,
			"foo",
			logEntry{
				Severity: "WARNING",
				Message:  "foo",
				SourceLocation: sourceLocation{
					File: "log/zap/stackdriver_test.go",
					Line: 21,
				},
			},
		},
		{
			log.SevError,
			"foo",
			logEntry{
				Severity: "ERROR",
				Message:  "foo",
				SourceLocation: sourceLocation{
					File: "log/zap/stackdriver_test.go",
					Line: 21,
				},
			},
		},
		{
			log.SevFatal,
			"foo",
			logEntry{
				Severity: "CRITICAL",
				Message:  "foo",
				SourceLocation: sourceLocation{
					File: "log/zap/stackdriver_test.go",
					Line: 21,
				},
			},
		},
	}

	for _, test := range tests {
		output := callLogger(test.sev, test.msg)

		var actual logEntry
		decoder := json.NewDecoder(output)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&actual); err != nil {
			t.Errorf("Error while deserializing log entry:\n(%v)", err)
		}

		if actual.Severity != test.expected.Severity {
			t.Errorf("Incorrect severity:\n(%v), want:\n(%v)", actual.Severity, test.expected.Severity)
		}

		if actual.Message != test.expected.Message {
			t.Errorf("Incorrect message:\n(%v), want:\n(%v)", actual.Message, test.expected.Message)
		}

		if !strings.HasSuffix(actual.SourceLocation.File, test.expected.SourceLocation.File) {
			t.Errorf("Incorrect file:\n(%v), want it to have a suffix:\n(%v)", actual.SourceLocation.File, test.expected.SourceLocation.File)
		}

		if actual.SourceLocation.Line != test.expected.SourceLocation.Line {
			t.Errorf("Incorrect line:\n(%v), want:\n(%v)", actual.SourceLocation.Line, test.expected.SourceLocation.Line)
		}
	}
}
