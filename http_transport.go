package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"sync"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

func NewBypassJA3Transport(helloID utls.ClientHelloID) *BypassJA3Transport {
	return &BypassJA3Transport{clientHello: helloID}
}

type BypassJA3Transport struct {
	tr1 http.Transport
	tr2 http2.Transport

	mu          sync.RWMutex
	clientHello utls.ClientHelloID
}

func (b *BypassJA3Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.URL.Scheme {
	case "https":
		return b.httpsRoundTrip(req)
	case "http":
		return b.tr1.RoundTrip(req)
	default:
		return nil, fmt.Errorf("unsupported scheme: %s", req.URL.Scheme)
	}
}

func (b *BypassJA3Transport) httpsRoundTrip(req *http.Request) (*http.Response, error) {
	port := req.URL.Port()
	if port == "" {
		port = "443"
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", req.URL.Host, port))
	if err != nil {
		return nil, fmt.Errorf("tcp net dial fail: %w", err)
	}
	defer conn.Close() // nolint

	tlsConn, err := b.tlsConnect(conn, req)
	if err != nil {
		return nil, fmt.Errorf("tls connect fail: %w", err)
	}

	httpVersion := tlsConn.ConnectionState().NegotiatedProtocol
	switch httpVersion {
	case "h2":
		conn, err := b.tr2.NewClientConn(tlsConn)
		if err != nil {
			return nil, fmt.Errorf("create http2 client with connection fail: %w", err)
		}
		defer conn.Close() // nolint
		return conn.RoundTrip(req)
	case "http/1.1", "":
		err := req.Write(tlsConn)
		if err != nil {
			return nil, fmt.Errorf("write http1 tls connection fail: %w", err)
		}
		return http.ReadResponse(bufio.NewReader(tlsConn), req)
	default:
		return nil, fmt.Errorf("unsuported http version: %s", httpVersion)
	}
}

func (b *BypassJA3Transport) getTLSConfig(req *http.Request) *utls.Config {
	return &utls.Config{
		ServerName:         req.URL.Host,
		InsecureSkipVerify: true,
	}
}

func (b *BypassJA3Transport) tlsConnect(conn net.Conn, req *http.Request) (*utls.UConn, error) {
	b.mu.RLock()
	tlsConn := utls.UClient(conn, b.getTLSConfig(req), b.clientHello)
	b.mu.RUnlock()

	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("tls handshake fail: %w", err)
	}
	return tlsConn, nil
}

func (b *BypassJA3Transport) SetClientHello(hello utls.ClientHelloID) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clientHello = hello
}
