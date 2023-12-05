# HTTP Tunnel Proxy

This HTTP tunnel proxy is designed to serve both HTTP/1.1 and HTTP/2 requests over TCP while randomizing its TLS fingerprint to evade scraper blocking through TLS fingerprinting techniques. 

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)

## Features

- **HTTP/1.1 and HTTP/2 Support:** This proxy supports both HTTP/1.1 and HTTP/2 protocols, making it versatile for various web applications.

- **Randomized TLS Fingerprint:** The proxy continuously randomizes its TLS fingerprint, making it difficult scraper detection to identify and block based on TLS fingerprinting.

- **TCP Tunneling:** The proxy tunnels traffic over TCP, ensuring robust and reliable communication.
## Source Installation

1. Clone this repository to your server:

   ```bash
   git clone https://github.com/Deathwalker9959/native-http-proxy.git
   ```

2. Navigate to the project directory:

   ```bash
   cd native-http-proxy
   ```

3. Install the required dependencies:

   ```bash
   go get
   ```
4. Install as binary
   ```bash
   go install native-http-proxy
   ```

## Usage

1. Start the HTTP tunnel proxy:

   ```bash
   ./native-http-proxy --port 8080
   ```

2. Configure your client to use the proxy. Ensure that you point your client to the correct proxy server address and port.

3. Enjoy seamless HTTP/1.1 and HTTP/2 traffic while evading scraper blocking!

## Configuration

You can customize the behavior of the HTTP tunnel proxy by modifying the commandline arguements. Here are some of the available configuration options:

- `port`: The port on which the proxy server listens.

Make sure to review and update the configuration to suit your specific requirements.

## Contributing

Contributions are welcome! If you want to contribute to this project, please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and commit them.
4. Create a pull request describing your changes.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Feel free to use this HTTP tunnel proxy for your web scraping, data gathering, or other network-related projects. If you have any questions or encounter issues, don't hesitate to open an issue or reach out to us for support. Happy proxying!
