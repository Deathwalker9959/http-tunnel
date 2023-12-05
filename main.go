package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/imroc/req/v3"
	utls "github.com/refraction-networking/utls"
)

var CERT, CERT_ERR = utls.LoadX509KeyPair("server-cert.pem", "server-key.pem")

func main() {
	if CERT_ERR != nil {
		log.Fatalf("Error loading X.509 key pair: %s", CERT_ERR)
	}

	go startHTTPProxy()
	go startHTTPServer()

	select {}
}

func startHTTPServer() {
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().HijackConnect(handleConnect)
	log.Fatal(http.ListenAndServe(":8081", proxy))
}

func handleConnect(req *http.Request, client net.Conn, ctx *goproxy.ProxyCtx) {
	serverName := req.URL.Hostname()

	client.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	uServer := utls.Server(client, &utls.Config{ServerName: serverName, InsecureSkipVerify: true, Certificates: []utls.Certificate{CERT}})
	if err := uServer.Handshake(); err != nil {
		log.Printf("Handshake error: %s", err)
		return
	}

	dialer, err := net.Dial("tcp", ":8443")
	if err != nil {
		log.Printf("net.Dial() error: %s", err)
		return
	}

	go pipeConnection(dialer, uServer)
	go pipeConnection(uServer, dialer)
}

func startHTTPProxy() {
	httpProxy, err := net.Listen("tcp", ":8443")
	if err != nil {
		log.Fatalf("Error listening on port 8443: %s", err)
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
