package sidecar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"

	"google.golang.org/protobuf/proto"
)

// Send writes a message to a writer with gRPC framing.
//
// The value must be a proto.Message; if not, an error is returned.
func Send(w io.Writer, value any) error {
	buf, err := serialize(value)
	if err != nil {
		return err
	}
	buf.WriteTo(w)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// Receive reads a value from a reader assuming gRPC framing.
//
// The value must be a proto.Message; if not, an error is returned.
func Receive(reader io.Reader, value any) error {
	// the first byte indicates compression, the next 4 are for message length
	prefix := make([]byte, 5)
	n, err := reader.Read(prefix)
	if err != nil {
		return err
	}
	if n != 5 {
		return io.EOF
	}
	compression := prefix[0]
	if compression != 0 {
		return fmt.Errorf("unsupported compression byte %d", compression)
	}
	length := binary.BigEndian.Uint32(prefix[1:5])
	body := make([]byte, length)
	n, err = reader.Read(body)
	if err != nil {
		return err
	}
	message, ok := value.(proto.Message)
	if !ok {
		return fmt.Errorf("invalid message type: %T", value)
	}
	return proto.Unmarshal(body, message)
}

func serialize(value any) (*bytes.Buffer, error) {
	message, ok := value.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("invalid message type: %T", value)
	}
	b, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(0) // no compression
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(b)))
	buf.Write(length)
	buf.Write(b)
	return &buf, nil
}
