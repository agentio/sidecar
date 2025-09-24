# echo-sidecar

This directory contains a command-line tool that includes a server and
clients for a sample gRPC service that demonstrates all four streaming
modes of gRPC.

The server and clients can communicate over TCP ports or Linux abstract sockets.

To use TCP ports, start the server with the `--port` flag:
```sh
echo-sidecar serve --port 8088
```

Call the client with the `--address` flag as below:
```sh
echo-sidecar call get --address localhost:8088
```

To use Linux abstract sockets, run the server with the `--socket` flag:
```sh
echo-sidecar serve --socket @echo
```

Call the client with the `--address` flag as below:
```sh
echo-sidecar call get --address unix:@echo
```

Running `echo-sidecar call` lists the four test methods:
```sh
$ echo-sidecar call
Usage:
  echo-sidecar call [command]

Available Commands:
  collect
  expand
  get
  update

Flags:
  -h, --help   help for call

Use "echo-sidecar call [command] --help" for more information about a command.
```

Running `go test` in this directory tests the server and clients for all four modes over both a local TCP connection and a Linux abstract socket.
```sh
$ go test . -v
=== RUN   TestSocket
--- PASS: TestSocket (0.02s)
=== RUN   TestLocal
--- PASS: TestLocal (0.02s)
PASS
ok      github.com/agentio/sidecar/cmd/echo-sidecar     0.043s
```

