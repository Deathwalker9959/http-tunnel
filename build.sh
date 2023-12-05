# Build for Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o ./dist/windows/native-http-proxy-amd64.exe  -v

GOOS=windows GOARCH=386 go build -o ./dist/windows/native-http-proxy-386.exe -v

# Build for Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o ./dist/linux/native-http-proxy-amd64 -v

# Build for Linux (32-bit)
GOOS=linux GOARCH=386 go build -o ./dist/linux/native-http-proxy-386 -v

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o ./dist/darwin/native-http-proxy-amd64 -v
