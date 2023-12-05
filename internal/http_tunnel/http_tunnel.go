package http_tunnel

import (
	"fmt"
	"io"
	"log"
	"native-http-proxy/internal/cert_mgmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/imroc/req/v3"
	utls "github.com/refraction-networking/utls"
)

var CERT, CERT_ERR = cert_mgmt.LoadX509KeyPair(CERT_PATH+"server.pem", CERT_PATH+"server.key")
var CERT_PATH = fmt.Sprint(os.TempDir(), "/native_http_proxy/")

var HTTPS_PORT, HTTPS_PORT_ERR = findAvailablePort()

func StartHTTPTunnel(port uint16) {

	log.Printf("Loading certificate from %s", fmt.Sprint(os.TempDir(), "/native_http_proxy/server.pem"))

	if CERT_ERR != nil {
		log.Fatalf("Failed to load certificate: %s", CERT_ERR)
	}

	if HTTPS_PORT_ERR != nil {
		log.Fatalf("Failed to find available port: %s", HTTPS_PORT_ERR)
	}

	log.Printf("Starting HTTP tunnel on port %d", port)
	log.Printf("Starting HTTPS proxy on port %d", HTTPS_PORT)

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().HijackConnect(handleConnect)

	go func() {
		log.Fatal(http.ListenAndServe(":"+fmt.Sprint(port), proxy))
	}()

	go startHTTPSProxy()

	select {}
}

func handleConnect(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
	serverName := req.URL.Hostname()

	client.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	uServer := utls.Server(client, &utls.Config{ServerName: serverName, InsecureSkipVerify: true, Certificates: []utls.Certificate{CERT}})
	if err := uServer.Handshake(); err != nil {
		log.Printf("Handshake error: %s", err)
		return
	}

	dialer, err := net.Dial("tcp", ":"+fmt.Sprint(HTTPS_PORT))
	if err != nil {
		log.Printf("net.Dial() error: %s", err)
		return
	}

	go pipeConnection(dialer, uServer)
	go pipeConnection(uServer, dialer)
}

func findAvailablePort() (uint16, error) {
	listener, err := net.Listen("tcp", "localhost:0") // Listen on a random available port
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return uint16(addr.Port), nil
}

func startHTTPSProxy() {
	httpProxy, err := net.Listen("tcp", ":"+fmt.Sprint(HTTPS_PORT))
	if err != nil {
		log.Fatalf("Error listening on port %d: %s", HTTPS_PORT, err)
	}
	for {
		conn, err := httpProxy.Accept()
		if err != nil {
			log.Printf("httpProxy.Accept() error: %s", err)
			return
		}
		go handleLocalTLS(conn)
		defer conn.Close()
	}
}

func handleLocalTLS(conn net.Conn) {
	buffer := make([]byte, 0)
	for {
		tmp := make([]byte, 1024)
		n, err := conn.Read(tmp)
		if err != nil {
			log.Printf("conn.Read() error: %s", err)
			return
		}
		buffer = append(buffer, tmp[:n]...)
		if n < 1024 {
			break
		}
	}

	bufferStr := string(buffer)
	headers, body, method, path, host := parseHTTP(bufferStr)

	client := req.C().SetAutoDecodeAllContentType().ImpersonateChrome().SetTLSFingerprintChrome()
	resp, err := client.R().SetHeaders(headers).SetBody(body).Send(method, "https://"+host+path)
	if err != nil {
		log.Printf("Request send error: %s", err)
		return
	}

	var respBuilder strings.Builder
	respBuilder.WriteString("HTTP/1.1 " + resp.Status + "\r\n")
	for k, v := range resp.Header {
		respBuilder.WriteString(k + ": " + strings.Join(v, ", ") + "\r\n")
	}
	respBuilder.WriteString("\r\n")
	respBuilder.Write(resp.Bytes())

	conn.Write([]byte(respBuilder.String()))
	conn.Close()
}

func pipeConnection(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()
	io.Copy(dst, src)
}

func parseHTTP(bufferStr string) (map[string]string, string, string, string, string) {
	// Split the request into header and body
	parts := strings.SplitN(bufferStr, "\r\n\r\n", 2)
	headerPart := parts[0]
	body := ""
	if len(parts) > 1 {
		body = parts[1]
	}

	// Split the header part into lines
	lines := strings.Split(headerPart, "\r\n")
	requestLine := lines[0]
	headerLines := lines[1:]

	// Parse the request line
	requestLineParts := strings.Split(requestLine, " ")
	method := requestLineParts[0]
	path := requestLineParts[1]

	// Initialize headers map
	headers := make(map[string]string)

	// Variable to hold the value of the Host header
	host := ""

	// Parse each header line
	for _, line := range headerLines {
		if len(line) == 0 {
			continue
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) != 2 {
			continue
		}
		headerName := headerParts[0]
		headerValue := headerParts[1]
		headers[headerName] = headerValue

		// Special handling for the Host header
		if headerName == "Host" {
			host = headerValue
		}
	}

	return headers, body, method, path, host
}
