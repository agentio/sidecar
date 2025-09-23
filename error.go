/*
 *
 * Copyright 2017 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package status implements errors returned by gRPC.  These errors are
// serialized and transmitted on the wire between server and client, and allow
// for additional data to be transmitted via the Details field in the status
// proto.  gRPC service handlers should return an error created by this
// package, and gRPC clients should expect a corresponding error to be
// returned from the RPC call.
//
// This package upholds the invariants that a non-nil error may not
// contain an OK code, and an OK code must result in a nil error.
package sidecar

import (
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
