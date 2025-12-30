package main

import (
	"encoding/json"
	"os"
)

const (
	settingsFile = "settings.json"
)

type Settings struct {
	RecentProjects []string           `json:"recentProjects"`
	Certificate    *CertificateConfig `json:"certificate,omitempty"`
}

func InitSettings() *Settings {
	var settings Settings
	// Load settings from file if exists
	if data, err := os.ReadFile(settingsFile); err == nil {
		err := json.Unmarshal(data, &settings)
		if err != nil {
			println("Error loading settings:", err.Error())
		}
	}
	return &settings
}

// addRecentProject adds a path to the recent projects list
func (settings *Settings) addRecentProject(path string) {
	// Remove if already exists
	for i, p := range settings.RecentProjects {
		if p == path {
			settings.RecentProjects = append(settings.RecentProjects[:i], settings.RecentProjects[i+1:]...)
			break
		}
	}
	// Add to front
	settings.RecentProjects = append([]string{path}, settings.RecentProjects...)
	// Keep max 10
	if len(settings.RecentProjects) > 10 {
		settings.RecentProjects = settings.RecentProjects[:10]
	}
	settings.save()
}

// save writes the settings to file
func (settings *Settings) save() {
	// Save to file
	data, err := json.MarshalIndent(settings, "", "  ")
	if err == nil {
		os.WriteFile(settingsFile, data, 0644)
	}
}
