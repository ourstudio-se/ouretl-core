package main

import ouretl "github.com/ourstudio-se/ouretl-abstractions"

type defaultPluginDefinition struct {
	NameVal         string `toml:"name"`
	PathVal         string `toml:"path"`
	VersionVal      string `toml:"version"`
	PriorityVal     int    `toml:"priority"`
	SettingsFileVal string `toml:"settings_file"`
	isActive        bool
	settings        *defaultPluginSettings
}

func (dpd *defaultPluginDefinition) Name() string {
	return dpd.NameVal
}

func (dpd *defaultPluginDefinition) FilePath() string {
	return dpd.PathVal
}

func (dpd *defaultPluginDefinition) Version() string {
	return dpd.VersionVal
}

func (dpd *defaultPluginDefinition) Priority() int {
	return dpd.PriorityVal
}

func (dpd *defaultPluginDefinition) Settings() ouretl.PluginSettings {
	return dpd.settings
}

func (dpd *defaultPluginDefinition) IsActive() bool {
	return dpd.isActive
}

type byPriority []*defaultPluginDefinition

func (w byPriority) Len() int {
	return len(w)
}

func (w byPriority) Swap(i, j int) {
	w[i], w[j] = w[j], w[i]
}

func (w byPriority) Less(i, j int) bool {
	return w[i].Priority() < w[j].Priority()
}
