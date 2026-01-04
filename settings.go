package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Settings struct {
	RecentProjects []string          `json:"recentProjects"`
	Certificate    CertificateConfig `json:"certificate"`
}

func settingsFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "settings.json"
	}
	appConfigDir := filepath.Join(configDir, "resttester")
	os.MkdirAll(appConfigDir, 0755)
	return filepath.Join(appConfigDir, "settings.json")
}
func InitSettings() (*Settings, error) {
	var settings Settings
	// Load settings from file if exists
	if data, err := os.ReadFile(settingsFilePath()); err == nil {
		err := json.Unmarshal(data, &settings)
		if err != nil {
			return nil, err
		}
	}
	return &settings, nil
}

// addRecentProject adds a path to the recent projects list
func (settings *Settings) addRecentProject(path string) error {
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
	return settings.save()
}

// save writes the settings to file
func (settings *Settings) save() error {
	// Save to file
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(settingsFilePath()), 0755); err != nil {
		return err
	}
	return os.WriteFile(settingsFilePath(), data, 0644)
}
