package codes

import (
	"net/http"
	"strconv"
)

func CodeFromResponse(resp *http.Response) Code {
	grpcstatus := resp.Header.Get("Grpc-Status")
	if grpcstatus != "" {
		code, err := strconv.Atoi(grpcstatus)
		if err == nil {
			return Code(code)
		} else {
			return MaxCode
		}
	}
	return OK
}
