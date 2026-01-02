package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CertificateConfig holds client certificate settings
type CertificateConfig struct {
	CertFile   string `json:"certFile"`   // Path to PEM certificate file
	KeyFile    string `json:"keyFile"`    // Path to PEM private key file
	CACertFile string `json:"caCertFile"` // Optional: Path to CA bundle file
	SkipVerify bool   `json:"skipVerify"` // Skip TLS certificate verification
}

type ProjectSettings struct {
	TimeoutInMs int64 `json:"timeoutInMs"` // Request timeout in milliseconds
}

// Project represents a collection of saved requests
type Project struct {
	Name     string          `json:"name"`
	Version  string          `json:"version"`
	Requests []*Request      `json:"requests"`
	Settings ProjectSettings `json:"settings"`
	filePath string          // Not saved, tracks where project is stored
}

// NewProject creates a new empty project
func NewProject(name string) *Project {
	return &Project{
		Name:     name,
		Version:  "1.0",
		Requests: make([]*Request, 0),
		Settings: ProjectSettings{
			TimeoutInMs: 30000, // Default 30 seconds
		},
	}
}

// NewRequest adds a new request to the project
func (p *Project) NewRequest() *Request {
	// Create a new request
	req := NewRequest(fmt.Sprintf("Request %d", len(p.Requests)+1))
	req.Headers["Content-Type"] = "application/json"
	req.Headers["Accept"] = "application/json"
	p.Requests = append(p.Requests, req)
	return req
}

// RemoveRequest removes a request by index
func (p *Project) RemoveRequest(index int) {
	if index < 0 || index >= len(p.Requests) {
		return
	}
	p.Requests = append(p.Requests[:index], p.Requests[index+1:]...)
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
