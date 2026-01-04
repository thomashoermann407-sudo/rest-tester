package main

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// NodeType represents the type of a tree node
type NodeType int

const (
	NodeTypeHost   NodeType = iota // Root level - host/domain
	NodeTypePath                   // Path segments
	NodeTypeMethod                 // HTTP method (leaf)
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

type ProjectSettings struct {
	TimeoutInMs int64 `json:"timeoutInMs"` // Request timeout in milliseconds
}

// RequestNode represents a node in the hierarchical REST resource tree
type RequestNode struct {
	Segment  string                  `json:"segment"`  // URL segment (e.g., "users", "api", "v1")
	Requests map[string]*Request     `json:"requests"` // Requests by method (GET, POST, etc.)
	Children map[string]*RequestNode `json:"children"` // Child nodes by segment name
}

// NewRequestNode creates a new request node
func NewRequestNode(segment string) *RequestNode {
	return &RequestNode{
		Segment:  segment,
		Requests: make(map[string]*Request),
		Children: make(map[string]*RequestNode),
	}
}

// AddRequest adds a request to this node for a specific HTTP method
func (n *RequestNode) AddRequest(method string, req *Request) {
	n.Requests[method] = req
}

// GetRequest retrieves a request for a specific HTTP method
func (n *RequestNode) GetRequest(method string) *Request {
	return n.Requests[method]
}

// GetOrCreateChild gets or creates a child node by segment name
func (n *RequestNode) GetOrCreateChild(segment string) *RequestNode {
	if child, exists := n.Children[segment]; exists {
		return child
	}
	child := NewRequestNode(segment)
	n.Children[segment] = child
	return child
}

// FindNode finds a node by following a path of segments
func (n *RequestNode) FindNode(segments []string) *RequestNode {
	if len(segments) == 0 {
		return n
	}
	if child, exists := n.Children[segments[0]]; exists {
		return child.FindNode(segments[1:])
	}
	return nil
}

// RequestTree represents the hierarchical organization of requests
type RequestTree struct {
	Root *RequestNode `json:"root"`
}

// NewRequestTree creates a new request tree
func NewRequestTree() *RequestTree {
	return &RequestTree{
		Root: NewRequestNode(""),
	}
}

// ParseURLPath parses a URL and extracts the path segments
func ParseURLPath(urlStr string) ([]string, error) { //TODO: remove
	if urlStr == "" {
		return []string{}, nil
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Get the path and split into segments
	path := strings.Trim(parsedURL.Path, "/")
	if path == "" {
		return []string{}, nil
	}

	segments := strings.Split(path, "/")
	return segments, nil
}

// AddRequest adds a request to the tree based on its URL
func (t *RequestTree) AddRequest(req *Request) error {
	segments, err := ParseURLPath(req.Path)
	if err != nil {
		return err
	}

	// Navigate/create the tree structure
	currentNode := t.Root
	for _, segment := range segments {
		currentNode = currentNode.GetOrCreateChild(segment)
	}

	// Add the request to the final node
	currentNode.AddRequest(req.Method, req)
	return nil
}

// RemoveRequest removes a request from the tree
func (t *RequestTree) RemoveRequest(req *Request) error {
	segments, err := ParseURLPath(req.Path)
	if err != nil {
		return err
	}

	node := t.Root.FindNode(segments)
	if node != nil {
		delete(node.Requests, req.Method)
	}
	return nil
}

// GetAllRequests returns a flat list of all requests in the tree
func (t *RequestTree) GetAllRequests() []*Request {
	var requests []*Request
	t.collectRequests(t.Root, &requests)
	return requests
}

func (t *RequestTree) collectRequests(node *RequestNode, requests *[]*Request) {
	// Add all requests from this node
	for _, req := range node.Requests {
		*requests = append(*requests, req)
	}

	// Recursively collect from children
	for _, child := range node.Children {
		t.collectRequests(child, requests)
	}
}

// Project represents a collection of saved requests
type Project struct {
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Tree         *RequestTree    `json:"tree"`
	Settings     ProjectSettings `json:"settings"`
	Environments []Environment   `json:"environments"` // Available environments
	filePath     string          // Not saved, tracks where project is stored
}

// NewProject creates a new empty project
func NewProject(name string) *Project {
	return &Project{
		Name:    name,
		Version: "2.0",
		Tree:    NewRequestTree(),
		Settings: ProjectSettings{
			TimeoutInMs: 30000, // Default 30 seconds
		},
		Environments: []Environment{
			{Name: "Local", BaseURL: "http://localhost:8080"},
		},
	}
}

// NewRequest adds a new request to the project
func (p *Project) NewRequest() *Request {
	// Create a new request
	req := NewRequest("Request")
	req.Headers["Content-Type"] = "application/json"
	req.Headers["Accept"] = "application/json"
	p.Tree.AddRequest(req)
	return req
}

// AddRequestToTree adds an existing request to the tree structure
func (p *Project) AddRequestToTree(req *Request) error {
	return p.Tree.AddRequest(req)
}

// RemoveRequestFromTree removes a request from the tree
func (p *Project) RemoveRequestFromTree(req *Request) error {
	// Remove from tree
	if err := p.Tree.RemoveRequest(req); err != nil {
		return err
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
