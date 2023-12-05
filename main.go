package main

import (
	"flag"

	"github.com/Deathwalker9959/native-http-proxy/internal/http_tunnel"
)

func main() {
	// Define the port flag with default value and description
	port := flag.Uint("port", 8080, "Port to listen on")

	// Parse command line flags
	flag.Parse()

	// Cast the port value to uint16
	castedPort := uint16(*port)

	// Start an HTTP tunnel in a goroutine
	go http_tunnel.StartHTTPTunnel(castedPort)

	// Block indefinitely to keep the program running
	select {}
}
