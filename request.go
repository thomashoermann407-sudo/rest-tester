package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"
)

type Params map[string]string

func (p Params) Format() string {
	var builder strings.Builder
	for _, key := range slices.Sorted(maps.Keys(p)) {
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(p[key])
		builder.WriteString("\r\n")
	}
	return builder.String()
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
	Host        string `json:"host"`
	Path        string `json:"path"`
	Headers     Params `json:"headers"`
	QueryParams Params `json:"queryParams"`
	Body        string `json:"body"`
}

// NewRequest creates a new request with default values
func NewRequest(name string) *Request {
	return &Request{
		Name:        name,
		Method:      "GET",
		Path:        "/",
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Body:        "",
	}
}

func (request *Request) sendRequest(settings *Settings, timeoutInMs int64, callback func(responseData *ResponseData, err error)) {
	startTime := time.Now()

	var requestUrl string
	if len(request.QueryParams) == 0 {
		requestUrl = request.Host + request.Path
	} else {
		v := url.Values{}
		for key, value := range request.QueryParams {
			v.Set(key, value)
		}
		requestUrl = request.Host + request.Path + "?" + v.Encode()
	}
	// Create request
	var reqBody io.Reader
	if request.Method == "POST" || request.Method == "PUT" || request.Method == "PATCH" {
		reqBody = strings.NewReader(request.Body)
	}

	req, err := http.NewRequest(request.Method, requestUrl, reqBody)
	if err != nil {
		callback(nil, err)
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

	// Create HTTP client with TLS configuration and timeout
	client, err := createHTTPClient(settings, timeoutInMs)
	if err != nil {
		callback(nil, err)
		return
	}

	resp, err := client.Do(req)
	duration := time.Since(startTime)

	if err != nil {
		callback(nil, err)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		callback(nil, err)
		return
	}

	// Get content type to determine formatting
	contentType := resp.Header.Get("Content-Type")
	formattedBody := formatResponse(string(respBody), contentType)

	// Extract response headers
	headers := make(map[string]string)
	for name, values := range resp.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	// Create response data
	responseData := &ResponseData{
		Body:       formattedBody,
		Headers:    headers,
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Duration:   duration,
		Timestamp:  time.Now(),
	}

	// Update UI
	callback(responseData, nil)
}

// createHTTPClient creates an HTTP client with optional TLS client certificate
func createHTTPClient(settings *Settings, timeoutInMs int64) (*http.Client, error) {
	client := &http.Client{
		Timeout: time.Duration(timeoutInMs) * time.Millisecond,
	}

	cert := &settings.Certificate
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
