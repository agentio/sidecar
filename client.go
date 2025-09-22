package sidecar

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"
)

// Client represents a gRPC client and includes an http.Client that
// is configured for H2C and reusable HTTP2 connections and the host
// name of a gRPC server. This is a minimal configuration for making
// gRPC calls.
type Client struct {
	httpclient *http.Client
	host       string
	Header     http.Header
}

// NewClient creates a client representation that includes an http.Client
// and a host name.
func NewClient(address string) *Client {
	// Configure protocols for H2C-only support (HTTP/2 cleartext)
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true) // Enable H2C (HTTP/2 cleartext)
	protocols.SetHTTP1(false)           // Explicitly disable HTTP/1.1
	protocols.SetHTTP2(false)           // Explicitly disable encrypted HTTP/2 (HTTPS)
	var client *http.Client
	if strings.HasPrefix(address, "unix:") {
		client = &http.Client{
			Transport: &http.Transport{
				Protocols: protocols,
				DialContext: func(ctx context.Context, _ string, addr string) (net.Conn, error) {
					network := "unix"
					addr = "@" + strings.TrimSuffix(strings.TrimPrefix(addr, "http://"), ":80")
					return net.DialTimeout(network, addr, 5*time.Second)
				},
			},
		}
	} else {
		client = &http.Client{
			Transport: &http.Transport{
				Protocols: protocols,
			},
		}
	}
	host := address
	if strings.HasPrefix(address, "unix:") {
		host = strings.Replace(address, "unix:", "http://", 1)
	} else if !strings.HasPrefix(address, "http://") {
		host = "http://" + address
	}
	return &Client{httpclient: client, host: host}
}

func (client *Client) addHeaders(request *http.Request) {
	for k, v := range client.Header {
		for _, vv := range v {
			request.Header.Add(k, vv)
		}
	}
}
