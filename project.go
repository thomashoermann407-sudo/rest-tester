package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Request represents a single HTTP request configuration
type Request struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	QueryParams map[string]string `json:"queryParams"`
	Body        string            `json:"body"`
}

// CertificateConfig holds client certificate settings
type CertificateConfig struct {
	CertFile   string `json:"certFile"`   // Path to PEM certificate file
	KeyFile    string `json:"keyFile"`    // Path to PEM private key file
	CACertFile string `json:"caCertFile"` // Optional: Path to CA bundle file
	SkipVerify bool   `json:"skipVerify"` // Skip TLS certificate verification
}

// Project represents a collection of saved requests
type Project struct {
	Name        string             `json:"name"`
	Version     string             `json:"version"`
	Certificate *CertificateConfig `json:"certificate,omitempty"`
	Requests    []*Request         `json:"requests"`
	filePath    string             // Not saved, tracks where project is stored
}

// NewProject creates a new empty project
func NewProject(name string) *Project {
	return &Project{
		Name:        name,
		Version:     "1.0",
		Certificate: &CertificateConfig{},
		Requests:    make([]*Request, 0),
	}
}

// NewRequest creates a new request with default values
func NewRequest(name string) *Request {
	return &Request{
		ID:          generateID(),
		Name:        name,
		Method:      "GET",
		URL:         "",
		Headers:     make(map[string]string),
		QueryParams: make(map[string]string),
		Body:        "",
	}
}

// AddRequest adds a new request to the project
func (p *Project) AddRequest(req *Request) {
	p.Requests = append(p.Requests, req)
}

// RemoveRequest removes a request by ID
func (p *Project) RemoveRequest(id string) {
	for i, req := range p.Requests {
		if req.ID == id {
			p.Requests = append(p.Requests[:i], p.Requests[i+1:]...)
			break
		}
	}
}

// GetRequest finds a request by ID
func (p *Project) GetRequest(id string) *Request {
	for _, req := range p.Requests {
		if req.ID == id {
			return req
		}
	}
	return nil
}

// GetRequestByIndex gets request at index
func (p *Project) GetRequestByIndex(index int) *Request {
	if index >= 0 && index < len(p.Requests) {
		return p.Requests[index]
	}
	return nil
}

// Save saves the project to a file
func (p *Project) Save(filePath string) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return err
	}

	p.filePath = filePath
	return nil
}

// Load loads a project from a file
func LoadProject(filePath string) (*Project, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, err
	}

	project.filePath = filePath
	return &project, nil
}

// FilePath returns the current file path
func (p *Project) FilePath() string {
	return p.filePath
}

// SetFilePath sets the file path
func (p *Project) SetFilePath(path string) {
	p.filePath = path
}

// HeadersToString converts headers map to string format
func (r *Request) HeadersToString() string {
	var lines []string
	for key, value := range r.Headers {
		lines = append(lines, key+": "+value)
	}
	return joinLines(lines)
}

// HeadersFromString parses string format to headers map
func (r *Request) HeadersFromString(s string) {
	r.Headers = make(map[string]string)
	lines := splitLines(s)
	for _, line := range lines {
		line = trimSpace(line)
		if line == "" {
			continue
		}
		parts := splitN(line, ":", 2)
		if len(parts) == 2 {
			r.Headers[trimSpace(parts[0])] = trimSpace(parts[1])
		}
	}
}

// QueryParamsToString converts query params map to string format
func (r *Request) QueryParamsToString() string {
	var lines []string
	for key, value := range r.QueryParams {
		lines = append(lines, key+"="+value)
	}
	return joinLines(lines)
}

// QueryParamsFromString parses string format to query params map
func (r *Request) QueryParamsFromString(s string) {
	r.QueryParams = make(map[string]string)
	lines := splitLines(s)
	for _, line := range lines {
		line = trimSpace(line)
		if line == "" {
			continue
		}
		parts := splitN(line, "=", 2)
		if len(parts) == 2 {
			r.QueryParams[trimSpace(parts[0])] = trimSpace(parts[1])
		}
	}
}

// BuildURLWithParams returns URL with query parameters appended
func (r *Request) BuildURLWithParams() string {
	if len(r.QueryParams) == 0 {
		return r.URL
	}

	url := r.URL
	separator := "?"
	if containsChar(url, '?') {
		separator = "&"
	}

	for key, value := range r.QueryParams {
		url += separator + key + "=" + value
		separator = "&"
	}

	return url
}

// Helper functions to avoid import in this file (will use strings package)
var (
	nextID = 1
)

func generateID() string {
	id := nextID
	nextID++
	return intToString(id)
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	return string(result)
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\r\n"
		}
		result += line
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else if c != '\r' {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitN(s string, sep string, n int) []string {
	var result []string
	remaining := s
	sepLen := len(sep)

	for i := 0; i < n-1 && len(remaining) > 0; i++ {
		idx := indexOf(remaining, sep)
		if idx < 0 {
			break
		}
		result = append(result, remaining[:idx])
		remaining = remaining[idx+sepLen:]
	}
	result = append(result, remaining)
	return result
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}

func containsChar(s string, c rune) bool {
	for _, ch := range s {
		if ch == c {
			return true
		}
	}
	return false
}
