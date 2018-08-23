package main

import (
	"os"

	"github.com/BurntSushi/toml"
)

type defaultPluginSettings struct {
	settings        map[string]interface{}
	overrideFromEnv bool
}

func (dps *defaultPluginSettings) Get(key string) (interface{}, bool) {
	if dps.overrideFromEnv && os.Getenv(key) != "" {
		return os.Getenv(key), true
	}

	value, ok := dps.settings[key]
	return value, ok
}

func readSettingsFromTOMLFile(settingsFilePath string) *defaultPluginSettings {
	if _, err := os.Stat(settingsFilePath); err != nil {
		return &defaultPluginSettings{
			settings: make(map[string]interface{}),
		}
	}

	var settings map[string]interface{}
	if _, err := toml.DecodeFile(settingsFilePath, settings); err != nil {
		return &defaultPluginSettings{
			settings: make(map[string]interface{}),
		}
	}

	return &defaultPluginSettings{
		settings: settings,
	}
}
