package main

import (
	"bytes"
	"io"
	"log"
	"testing"
	"time"

	"github.com/agentio/sidecar/cmd/echo-sidecar/commands"
)

func TestSocket(t *testing.T) {
	test_service(t,
		[]string{"serve", "--socket", "@echotest"},
		[]string{"--address", "unix:@echotest"},
	)
}

func TestLocal(t *testing.T) {
	const port = "19872"
	test_service(t,
		[]string{"serve", "--port", port},
		[]string{"--address", "localhost:" + port},
	)
}

func test_service(t *testing.T, serverArgs, clientArgs []string) {
	go func() {
		serveCmd := commands.Cmd()
		serveCmd.SetArgs(serverArgs)
		err := serveCmd.Execute()
		if err != nil {
			log.Printf("failed to read output from buffer: %v", err)
		}
	}()
	time.Sleep(10 * time.Millisecond)
	tests := []struct {
		Args     []string
		Expected string
	}{
		{
			Args:     []string{"call", "get"},
			Expected: expected_get,
		},
		{
			Args:     []string{"call", "collect"},
			Expected: expected_collect,
		},
		{
			Args:     []string{"call", "expand"},
			Expected: expected_expand,
		},
		{
			Args:     []string{"call", "update"},
			Expected: expected_update,
		},
	}
	for _, test := range tests {
		cmd := commands.Cmd()
		buffer := new(bytes.Buffer)
		cmd.SetOut(buffer)
		cmd.SetArgs(append(test.Args, clientArgs...))
		err := cmd.Execute()
		if err != nil {
			t.Errorf("%s", err)
		}
		out, err := io.ReadAll(buffer)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}
		if string(out) != test.Expected {
			t.Errorf("expected %q, got %q", test.Expected, string(out))
		}
	}
}

const expected_get = `{"text":"Go echo get: hello"}
`
const expected_collect = `{"text":"Go echo collect: hello hello hello"}
`
const expected_expand = `{"text":"Go echo expand: 1"}
{"text":"Go echo expand: 2"}
{"text":"Go echo expand: 3"}
`
const expected_update = `{"text":"Go echo update: hello"}
{"text":"Go echo update: hello"}
{"text":"Go echo update: hello"}
{"text":"Go echo update: hello"}
{"text":"Go echo update: hello"}
{"text":"Go echo update: hello"}
`
