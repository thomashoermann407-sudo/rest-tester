package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
)

// formatResponse formats the response body based on content type
func formatResponse(body string, contentType string) string {
	contentType = strings.ToLower(contentType)

	// Try JSON formatting
	if strings.Contains(contentType, "json") || strings.HasPrefix(strings.TrimSpace(body), "{") || strings.HasPrefix(strings.TrimSpace(body), "[") {
		if formatted, ok := formatJSON(body); ok {
			return formatted
		}
	}

	// Try XML formatting
	if strings.Contains(contentType, "xml") || strings.HasPrefix(strings.TrimSpace(body), "<") {
		if formatted, ok := formatXML(body); ok {
			return formatted
		}
	}

	// Plain text - just return as-is with Windows line endings
	return strings.ReplaceAll(body, "\n", "\r\n")
}

// formatJSON pretty-prints JSON
func formatJSON(input string) (string, bool) {
	var data any
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return "", false
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", false
	}

	// Convert to Windows line endings
	return strings.ReplaceAll(string(formatted), "\n", "\r\n"), true
}

// formatXML pretty-prints XML
func formatXML(input string) (string, bool) {
	var buf strings.Builder
	decoder := xml.NewDecoder(strings.NewReader(input))
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", false
		}
		if err := encoder.EncodeToken(token); err != nil {
			return "", false
		}
	}

	if err := encoder.Flush(); err != nil {
		return "", false
	}

	// Convert to Windows line endings
	return strings.ReplaceAll(buf.String(), "\n", "\r\n"), true
}
