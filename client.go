package sidecar

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

// Client represents a gRPC client and includes an http.Client,
// a host name, and a header to be sent with all requests.
type Client struct {
	Host       string
	Header     http.Header
	HttpClient *http.Client
}

// NewClient creates a client representation from an address.
// Addresses must be in the format "HOSTNAME:PORT" or "unix:@SOCKET".
// Connections to port 443 use TLS. All others are cleartext (h2c).
func NewClient(address string) *Client {
	// Expect TLS on port 443 and use the default HTTP client.
	if strings.HasSuffix(address, ":443") {
		return &Client{
			Host:       "https://" + address,
			Header:     defaultHeader(),
			HttpClient: http.DefaultClient,
		}
	}
	// All other clients need h2c-only support (HTTP/2 cleartext).
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true) // Enable h2c (HTTP/2 cleartext)
	protocols.SetHTTP1(false)           // Explicitly disable HTTP/1.1
	protocols.SetHTTP2(false)           // Explicitly disable encrypted HTTP/2 (HTTPS)
	// If required, create a client that can call unix sockets.
	if strings.HasPrefix(address, "unix:") {
		return &Client{
			Host:   strings.Replace(address, "unix:", "http://", 1),
			Header: defaultHeader(),
			HttpClient: &http.Client{
				Transport: &http.Transport{
					Protocols: protocols,
					DialContext: func(ctx context.Context, _ string, addr string) (net.Conn, error) {
						addr = strings.TrimPrefix(addr, "http://")
						addr = strings.TrimSuffix(addr, ":80")
						addr = "@" + addr
						return net.DialTimeout("unix", addr, 5*time.Second)
					},
				},
			},
		}
	}
	// Create a client for networked h2c connections.
	return &Client{
		Host:   "http://" + address,
		Header: defaultHeader(),
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Protocols: protocols,
			},
		},
	}
}

func defaultHeader() http.Header {
	var header http.Header
	header = make(map[string][]string)
	header.Set("Content-Type", "application/grpc")
	return header
}
