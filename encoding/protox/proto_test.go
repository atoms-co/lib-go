package protox_test

import (
	"strings"
	"testing"

	legacyproto "github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.atoms.co/lib/encoding/protox"
	"go.atoms.co/lib/testing/requirex"
)

func TestUnmarshal(t *testing.T) {
	buf, err := proto.Marshal(&timestamppb.Timestamp{
		Seconds: 123,
		Nanos:   456,
	})
	require.NoError(t, err)

	ts, err := protox.Unmarshal[timestamppb.Timestamp](buf)
	require.NoError(t, err)
	requirex.Equal(t, ts.Seconds, 123)
	requirex.Equal(t, ts.Nanos, 456)
}

func TestUnmarshalLegacy(t *testing.T) {
	buf, err := legacyproto.Marshal(&timestamppb.Timestamp{
		Seconds: 123,
		Nanos:   456,
	})
	require.NoError(t, err)

	ts, err := protox.Unmarshal[timestamppb.Timestamp](buf)
	require.NoError(t, err)
	requirex.Equal(t, ts.Seconds, 123)
	requirex.Equal(t, ts.Nanos, 456)
}

func TestCompactTextString(t *testing.T) {
	pb := &timestamppb.Timestamp{
		Seconds: 123,
		Nanos:   456,
	}

	str := protox.CompactTextString(pb)
	// Need to reduce double whitespaces, prototext.encode adds randomly whitespace between fields
	requirex.Equal(t, strings.Replace(str, "  ", " ", -1), "seconds:123 nanos:456")
}

func TestMarshalTextString(t *testing.T) {
	pb := &timestamppb.Timestamp{
		Seconds: 123,
		Nanos:   456,
	}

	str := protox.MarshalTextString(pb)
	// Need to reduce double whitespaces, prototext.encode adds randomly whitespace between fields
	requirex.Equal(t, strings.Replace(str, "  ", " ", -1), `seconds: 123
nanos: 456
`)
}

func TestClone(t *testing.T) {
	pb := &timestamppb.Timestamp{
		Seconds: 123,
		Nanos:   456,
	}

	ts := protox.Clone(pb)
	requirex.Equal(t, ts.Seconds, 123)
	requirex.Equal(t, ts.Nanos, 456)
}
