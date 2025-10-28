package sidecar

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/agentio/sidecar/codes"
	"golang.org/x/net/http2"

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
	_, err = buf.WriteTo(w)
	if err != nil {
		return err
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

// Receive reads a value from a reader assuming gRPC framing.
//
// The value must be a proto.Message; if not, an error is returned.
func Receive(reader io.Reader, value any) error {
	b, err := unframe(reader)
	if errors.Is(err, io.EOF) {
		return err
	} else if err != nil {
		return NewError(err, codes.InvalidArgument)
	}
	// A []byte value is set to the raw message body.
	if byteSlice, ok := value.(*[]byte); ok {
		*byteSlice = b
		return nil
	}
	// A proto.Message value is set to the unmarshalled message.
	if message, ok := value.(proto.Message); ok {
		err = proto.Unmarshal(b, message)
		if err != nil {
			return fmt.Errorf("failed to unmarshal %T, %s", message, err)
		}
		return nil
	}
	return NewError(fmt.Errorf("unsupported message type: %T", value), codes.InvalidArgument)
}

func serialize(value any) (*bytes.Buffer, error) {
	// A []byte value is just wrapped in gRPC framing.
	if b, ok := value.(*[]byte); ok {
		return frame(*b), nil

	}
	// A proto.Message value is marshalled and framed.
	if message, ok := value.(proto.Message); ok {
		b, err := proto.Marshal(message)
		if err != nil {
			return nil, err
		}
		return frame(b), nil
	}
	return nil, NewError(fmt.Errorf("unsupported message type: %T", value), codes.InvalidArgument)
}

func frame(b []byte) *bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteByte(0) // no compression
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(b)))
	buf.Write(length)
	buf.Write(b)
	return &buf
}

func unframe(reader io.Reader) ([]byte, error) {
	// the first byte indicates compression, the next 4 are for message length
	prefix := make([]byte, 5)
	n, err := reader.Read(prefix)
	if err != nil {
		var streamErr http2.StreamError
		if errors.As(err, &streamErr) {
			if streamErr.Code == http2.ErrCodeNo { // end of stream, no error
				return nil, io.EOF
			}
		}
		return nil, err
	}
	if n != 5 {
		return nil, io.EOF
	}
	compression := prefix[0]
	if compression != 0 {
		return nil, fmt.Errorf("unsupported compression byte %d", compression)
	}
	length := binary.BigEndian.Uint32(prefix[1:5])
	b := make([]byte, length)
	n, err = io.ReadFull(reader, b)
	if err != nil {
		return nil, err
	}
	if n != int(length) {
		return nil, fmt.Errorf("failed to read %d bytes, got %d instead", length, n)
	}
	return b, err
}
