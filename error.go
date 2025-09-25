package sidecar

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/agentio/sidecar/codes"
)

type Error struct {
	err  error
	code codes.Code
}

func NewError(err error, code codes.Code) *Error {
	return &Error{
		err: err, code: code,
	}
}

func (s Error) Error() string {
	return s.err.Error()
}

func (s Error) Code() codes.Code {
	return s.code
}

func (s Error) Unwrap() error {
	return s.err
}

func ErrorCode(err error) int {
	if err == nil {
		return int(codes.OK)
	}
	code := codes.Internal
	if e, ok := err.(Error); ok {
		code = e.Code()
	}
	return int(code)
}

func WriteTrailer(w http.ResponseWriter, err error) {
	if err == nil {
		w.Header().Set("Trailer:Grpc-Status", strconv.Itoa(0))
		return
	}
	w.Header().Set("Trailer:Grpc-Status", strconv.Itoa(ErrorCode(err)))
	w.Header().Set("Trailer:Grpc-Message", err.Error())
}

func ErrorForTrailer(trailer http.Header) error {
	status := trailer.Get("Grpc-Status")
	if status == "0" {
		return nil
	}
	message := trailer.Get("Grpc-Message")
	code, _ := strconv.Atoi(status)
	return NewError(errors.New(message), codes.Code(code))
}
