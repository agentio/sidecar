build:
	go install ./...

all:	rpc
	go install ./...

clean:
	go clean
	rm -rf cmd/echo-sidecar/genproto

test:
	go test ./... -v

APIS=$(shell find cmd/echo-sidecar/proto/echo -name "*.proto")

descriptor:
	protoc ${APIS} \
	--proto_path='cmd/echo-sidecar/proto' \
	--include_imports \
	--descriptor_set_out=descriptor.pb

rpc:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	mkdir -p cmd/echo-sidecar/genproto
	protoc ${APIS} \
	--proto_path='cmd/echo-sidecar/proto' \
	--go_opt='module=github.com/agentio/sidecar/cmd/echo-sidecar/genproto' \
        --go_opt=Mecho/v1/echo.proto=github.com/agentio/sidecar/cmd/echo-sidecar/genproto/echopb \
	--go_out='cmd/echo-sidecar/genproto'

