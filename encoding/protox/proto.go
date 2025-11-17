package protox

import (
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/runtime/protoimpl"
)

// Unmarshal is a generic, direct-style wrapper for proto.Unmarshal that returns the message value directly.
//
// Type Parameters:
//
//	M - the concrete protobuf message type
//	T - an interface to satisfy type system, should not be provided
//
// This function differs from the standard proto.Unmarshal in that it does not require the caller to provide a pointer
// to the message. Instead, it creates a new instance of the message type, unmarshals the data into it, and returns
// a pointer to the message.
//
// Example usage:
//
//	msg, err := protox.Unmarshal[com.example.v1.MyMessage](data)
//	// msg is of type *com.example.v1.MyMessage
func Unmarshal[M any, T interface {
	*M
	proto.Message
}](buf []byte) (*M, error) {
	ret := new(M)
	if err := proto.Unmarshal(buf, protoimpl.X.ProtoMessageV2Of(ret)); err != nil {
		return nil, err
	}
	return ret, nil
}

// Marshal returns the wire-format encoding of m.
func Marshal(m proto.Message) ([]byte, error) {
	return proto.Marshal(m)
}

// CompactTextString returns a compact, single-line text representation of the given proto.Message.
// This format omits unnecessary whitespace and is suitable for logging or debugging where space is a concern.
func CompactTextString(m proto.Message) string {
	return prototext.MarshalOptions{}.Format(m)
}

// MarshalTextString returns a multiline, human-readable text representation of the given proto.Message.
// This format includes indentation and line breaks, making it easier to read for humans.
func MarshalTextString(m proto.Message) string {
	return prototext.Format(m)
}

// Clone returns a deep copy of m. If the top-level message is invalid,
// it returns an invalid message as well.
func Clone[M proto.Message](m M) M {
	return proto.Clone(m).(M)
}

// Equal reports whether two messages are equal,
// by recursively comparing the fields of the message.
func Equal(x, y proto.Message) bool {
	return proto.Equal(x, y)
}

// EqualList compares two message slices with proto.Equal. Convenience method.
func EqualList[M proto.Message](a, b []M) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if !proto.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
