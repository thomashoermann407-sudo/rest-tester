package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Params map[string]string

func (p Params) Format() string {
	var parts []string
	for key, value := range p {
		parts = append(parts, fmt.Sprintf("%s: %s", key, value))
	}
	return strings.Join(parts, "\r\n")
}
func ParseParams(input string) Params {
	params := make(Params)
	lines := strings.SplitSeq(input, "\r\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			params[name] = value
		}
	}
	return params
}

// Request represents a single HTTP request configuration
type Request struct {
	Name        string `json:"name"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	Headers     Params `json:"headers"`
	QueryParams Params `json:"queryParams"`
	Body        string `json:"body"`
}

// NewRequest creates a new request with default values
func NewRequest(name string) *Request {
	return &Request{
		Name:        name,
		Method:      "GET",
		URL:         "",
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Body:        "",
	}
}

func sendRequest(request *Request, settings *Settings, callback func(response string, err error)) {

	v := url.Values{}
	for key, value := range request.QueryParams {
		v.Set(key, value)
	}
	url := buildURLWithQueryParams(request.URL, v.Encode())

	// Create request
	var reqBody io.Reader
	if request.Method == "POST" || request.Method == "PUT" || request.Method == "PATCH" {
		reqBody = strings.NewReader(request.Body)
	}

	req, err := http.NewRequest(request.Method, url, reqBody)
	if err != nil {
		callback("", err)
		return
	}

	// Add headers from the map
	for name, value := range request.Headers {
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)
		if name != "" {
			req.Header.Set(name, value)
		}
	}

	// Create HTTP client with TLS configuration
	client, err := createHTTPClient(settings)
	if err != nil {
		callback("", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		callback("", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		callback("", err)
		return
	}

	// Get content type to determine formatting
	contentType := resp.Header.Get("Content-Type")
	formattedBody := formatResponse(string(respBody), contentType)

	// Update UI
	callback(formattedBody, nil)
}

// createHTTPClient creates an HTTP client with optional TLS client certificate
func createHTTPClient(settings *Settings) (*http.Client, error) {
	client := &http.Client{}

	// Check if we have certificate configuration
	if settings.Certificate == nil {
		return client, nil
	}

	cert := settings.Certificate
	certFile := strings.TrimSpace(cert.CertFile)
	keyFile := strings.TrimSpace(cert.KeyFile)

	// No certificate configured
	if certFile == "" && keyFile == "" && !cert.SkipVerify {
		return client, nil
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{}

	// Load client certificate if provided
	if certFile != "" && keyFile != "" {
		clientCert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	// Load CA certificate if provided
	if cert.CACertFile != "" {
		caCert, err := os.ReadFile(cert.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %v", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Skip TLS verification if requested
	tlsConfig.InsecureSkipVerify = cert.SkipVerify

	// Create transport with TLS config
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client.Transport = transport
	return client, nil
}
