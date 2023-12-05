package http_tunnel

import (
	"fmt"
	"testing"
	"time"

	"github.com/imroc/req/v3"
)

func TestHTTPTunnel(t *testing.T) {
	// Start the HTTP tunnel in a goroutine.
	port, err := findAvailablePort()
	if err != nil {
		t.Fatalf("Failed to find available port: %v", err)
	}

	go StartHTTPTunnel(port)
	// Allow some time for the tunnel to start.
	time.Sleep(time.Second)

	// Define the test cases.
	tests := []struct {
		httpVersion string
		name        string
		url         string
		useTLS      bool
		wantErr     bool
	}{
		{"HTTP1.1", "HTTP to httpbin.org", "http://httpbin.org/get", false, false},
		{"HTTP1.1", "HTTPS to httpbin.org", "https://httpbin.org/get", true, false},
		{"H2", "HTTPS to nghttp2.org", "https://nghttp2.org/httpbin/get", true, false},
	}

	// Run the test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := req.C().SetProxyURL("http://localhost:" + fmt.Sprint(port)).Get(tt.url)
			client.GetClient().EnableInsecureSkipVerify()
			switch tt.httpVersion {
			case "HTTP1.1":
				client.SetHeader("Connection", "close")
				client.GetClient().EnableForceHTTP1()
			case "H2":
				client.SetHeader("Connection", "keep-alive")
				client.GetClient().EnableForceHTTP2()
			}

			resp, err := client.Get(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTP tunnel test failed: %v", err)
			}

			if resp.StatusCode != 200 {
				t.Errorf("HTTP tunnel test failed: %v", resp.Status)
			}
		})
	}
}
