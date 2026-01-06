package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NodeType represents the type of a tree node
type NodeType int

const (
	NodeTypePath    NodeType = iota // Path segments
	NodeTypeRequest                 // Request nodes
)

// CertificateConfig holds client certificate settings
type CertificateConfig struct {
	CertFile   string `json:"certFile"`   // Path to PEM certificate file
	KeyFile    string `json:"keyFile"`    // Path to PEM private key file
	CACertFile string `json:"caCertFile"` // Optional: Path to CA bundle file
	SkipVerify bool   `json:"skipVerify"` // Skip TLS certificate verification
}

// Environment represents a deployment environment (dev, staging, prod, etc.)
type Environment struct {
	Name    string `json:"name"`    // Display name (e.g., "Development", "Production")
	BaseURL string `json:"baseUrl"` // Base URL including protocol and host (e.g., "https://api.example.com")
}

func (e *Environment) String() string {
	if e.Name != "" {
		return fmt.Sprintf("%s [%s]", e.Name, e.BaseURL)
	}
	return e.BaseURL
}

type ProjectSettings struct {
	TimeoutInMs           int64 `json:"timeoutInMs"`           // Request timeout in milliseconds
	DefaultEnvironmentIdx int   `json:"defaultEnvironmentIdx"` // Index of default environment (-1 for none)
}

// RequestNode represents a node in the hierarchical REST resource tree
type RequestNode struct {
	Segment  string         `json:"segment"`  // URL segment (e.g., "users", "api", "v1")
	Requests []*Request     `json:"requests"` // Requests by method (GET, POST, etc.)
	Children []*RequestNode `json:"children"` // Child nodes
}

// NewRequestNode creates a new request node
func NewRequestNode(segment string) *RequestNode {
	return &RequestNode{
		Segment: segment,
	}
}

// AddRequestAtPath adds a request at the specified full path
func (n *RequestNode) AddRequestAtPath(fullPath string, req *Request) {
	path := strings.TrimPrefix(fullPath, "/")
	if path == "" {
		// Adding to root node
		n.Requests = append(n.Requests, req)
		return
	}

	// Split into first segment and rest
	parts := strings.SplitN(path, "/", 2)
	firstSegment := parts[0]
	var remainingPath string
	if len(parts) > 1 {
		remainingPath = parts[1]
	}

	// Check if a child with this segment already exists
	for _, child := range n.Children {
		if child.Segment == firstSegment {
			// Found existing path segment, recurse into it
			if remainingPath == "" {
				// This is the final segment, add request here
				child.Requests = append(child.Requests, req)
			} else {
				// More path segments to process
				child.AddRequestAtPath(remainingPath, req)
			}
			return
		}
	}

	// No matching child found, create new path segments
	segments := strings.Split(path, "/")
	currentNode := n
	for _, segment := range segments {
		if segment == "" {
			continue
		}
		// Check if this segment already exists
		var found *RequestNode
		for _, child := range currentNode.Children {
			if child.Segment == segment {
				found = child
				break
			}
		}
		if found != nil {
			currentNode = found
		} else {
			newNode := NewRequestNode(segment)
			currentNode.Children = append(currentNode.Children, newNode)
			currentNode = newNode
		}
	}
	currentNode.Requests = append(currentNode.Requests, req)
}

// GetAllRequests returns a flat list of all requests in the tree
func (node *RequestNode) GetAllRequests() []*Request {
	var requests []*Request
	node.collectRequests(&requests)
	return requests
}

func (node *RequestNode) collectRequests(requests *[]*Request) {
	// Add all requests from this node
	for _, req := range node.Requests {
		*requests = append(*requests, req)
	}

	// Recursively collect from children
	for _, child := range node.Children {
		child.collectRequests(requests)
	}
}

// Project represents a collection of saved requests
type Project struct {
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Tree         *RequestNode    `json:"tree"`
	Settings     ProjectSettings `json:"settings"`
	Environments []Environment   `json:"environments"` // Available environments
	filePath     string          // Not saved, tracks where project is stored
}

// NewProject creates a new empty project
func NewProject(name string) *Project {
	return &Project{
		Name:    name,
		Version: "2.0",
		Tree:    NewRequestNode("/"),
		Settings: ProjectSettings{
			TimeoutInMs:           30000, // Default 30 seconds
			DefaultEnvironmentIdx: -1,    // No default environment
		},
		Environments: []Environment{
			{Name: "Local", BaseURL: "http://localhost:8080"},
		},
	}
}

// AddRequestToTree adds an existing request to the tree structure
func (p *Project) AddRequestToTree(fullPath string, req *Request) {
	p.Tree.AddRequestAtPath(fullPath, req)
}

func (p *Project) getDefaultHost() string {
	if p.Settings.DefaultEnvironmentIdx >= 0 && p.Settings.DefaultEnvironmentIdx < len(p.Environments) {
		env := p.Environments[p.Settings.DefaultEnvironmentIdx]
		return env.BaseURL
	}
	return ""
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
