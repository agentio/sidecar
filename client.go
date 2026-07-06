package sidecar

import (
	"context"
	"crypto/tls"
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

type ClientOptions struct {
	Address  string
	Insecure bool
	Headers  []string
}

// NewClient creates a client representation from an address.
// Addresses must be in the format "HOSTNAME:PORT" or "unix:@SOCKET".
// Connections to port 443 use TLS. All others are cleartext (h2c).
func NewClient(options ClientOptions) *Client {
	// Expect TLS on port 443 and force TLS clients to use HTTP2.
	if strings.HasSuffix(options.Address, ":443") {
		protocols := new(http.Protocols)
		protocols.SetUnencryptedHTTP2(false)
		protocols.SetHTTP1(false)
		protocols.SetHTTP2(true)
		return (&Client{
			Host:   "https://" + options.Address,
			Header: defaultHeader(),
			HttpClient: &http.Client{
				Transport: &http.Transport{
					Protocols:       protocols,
					TLSClientConfig: &tls.Config{InsecureSkipVerify: options.Insecure},
				},
			},
		}).addHeaders(options.Headers)
	}
	// All other clients need h2c-only support (HTTP/2 cleartext).
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true) // Enable h2c (HTTP/2 cleartext)
	protocols.SetHTTP1(false)           // Explicitly disable HTTP/1.1
	protocols.SetHTTP2(false)           // Explicitly disable encrypted HTTP/2 (HTTPS)
	// If required, create a client that can call unix sockets.
	if strings.HasPrefix(options.Address, "unix:") {
		address := strings.TrimPrefix(options.Address, "unix:")
		return (&Client{
			Host:   "http://socket", // The name "socket" is arbitrary.
			Header: defaultHeader(),
			HttpClient: &http.Client{
				Transport: &http.Transport{
					Protocols: protocols,
					DialContext: func(ctx context.Context, _ string, _ string) (net.Conn, error) {
						return net.DialTimeout("unix", address, 5*time.Second)
					},
				},
			},
		}).addHeaders(options.Headers)
	}
	// Create a client for networked h2c connections.
	return (&Client{
		Host:   "http://" + options.Address,
		Header: defaultHeader(),
		HttpClient: &http.Client{
			Transport: &http.Transport{
				Protocols: protocols,
			},
		},
	}).addHeaders(options.Headers)
}

func (c *Client) Close() {
	c.HttpClient.Transport.(*http.Transport).CloseIdleConnections()
}

func defaultHeader() http.Header {
	header := make(http.Header)
	header.Set("Content-Type", "application/grpc")
	header.Set("TE", "trailers")
	return header
}

func (client *Client) addHeaders(headers []string) *Client {
	for _, h := range headers {
		parts := strings.Split(h, ":")
		if len(parts) == 2 {
			client.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	return client
}
