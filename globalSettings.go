package main

import (
	"encoding/json"
	"os"
)

type GlobalSettings struct {
	RecentProjects []string           `json:"recentProjects"`
	Certificate    *CertificateConfig `json:"certificate,omitempty"`
}

var globalSettings GlobalSettings

func InitGlobalSettings() {
	// Load global settings from file if exists
	if data, err := os.ReadFile("globalSettings.json"); err == nil {
		err := json.Unmarshal(data, &globalSettings)
		if err != nil {
			println("Error loading global settings:", err.Error())
		}
	}
}

// addRecentProject adds a path to the recent projects list
func (gs *GlobalSettings) addRecentProject(path string) {
	// Remove if already exists
	for i, p := range gs.RecentProjects {
		if p == path {
			gs.RecentProjects = append(gs.RecentProjects[:i], gs.RecentProjects[i+1:]...)
			break
		}
	}
	// Add to front
	gs.RecentProjects = append([]string{path}, gs.RecentProjects...)
	// Keep max 10
	if len(gs.RecentProjects) > 10 {
		gs.RecentProjects = gs.RecentProjects[:10]
	}
	// Save to file
	data, err := json.MarshalIndent(gs, "", "  ")
	if err == nil {
		os.WriteFile("globalSettings.json", data, 0644)
	}
}
