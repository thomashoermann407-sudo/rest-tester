package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
)

// CertificateConfig holds client certificate settings
type CertificateConfig struct {
	CertFile   string `json:"certFile"`   // Path to PEM certificate file
	KeyFile    string `json:"keyFile"`    // Path to PEM private key file
	CACertFile string `json:"caCertFile"` // Optional: Path to CA bundle file
	SkipVerify bool   `json:"skipVerify"` // Skip TLS certificate verification
}

// Project represents a collection of saved requests
type Project struct {
	Name     string     `json:"name"`
	Version  string     `json:"version"`
	Requests []*Request `json:"requests"`
	filePath string     // Not saved, tracks where project is stored
}

// NewProject creates a new empty project
func NewProject(name string) *Project {
	return &Project{
		Name:     name,
		Version:  "1.0",
		Requests: make([]*Request, 0),
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

// Helper functions to avoid import in this file (will use strings package)
var (
	nextID int64 = 0
)

func generateID() string {
	id := atomic.AddInt64(&nextID, 1)
	return strconv.FormatInt(id, 10)
}
