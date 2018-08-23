package main

import (
	"os"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/radovskyb/watcher"
	log "github.com/sirupsen/logrus"

	ouretl "github.com/ourstudio-se/ouretl-abstractions"
)

type pluginDefinitionStatus int

var (
	pluginDefinitionActive   pluginDefinitionStatus = 10
	pluginDefinitionInactive pluginDefinitionStatus = 20
	pluginDefinitionMissing  pluginDefinitionStatus = 30
)

type defaultConfig struct {
	OverrideSettingsFromEnv     bool                       `toml:"inherit_settings_from_env"`
	Definitions                 []*defaultPluginDefinition `toml:"plugin"`
	onAddChangeListeners        []func(ouretl.PluginDefinition)
	onActivateChangeListeners   []func(ouretl.PluginDefinition)
	onDeactivateChangeListeners []func(ouretl.PluginDefinition)
}

func newDefaultConfig() ouretl.Config {
	var definitions []*defaultPluginDefinition
	return &defaultConfig{Definitions: definitions}
}

func newDefaultConfigFromTOMLFile(configFilePath string) (ouretl.Config, error) {
	config, err := readConfigFromFile(configFilePath)
	if err != nil {
		return nil, err
	}

	go config.createFileWatch(configFilePath)

	return config, nil
}

func readConfigFromFile(configFilePath string) (*defaultConfig, error) {
	if _, err := os.Stat(configFilePath); err != nil {
		return nil, err
	}

	var config defaultConfig
	if _, err := toml.DecodeFile(configFilePath, &config); err != nil {
		return nil, err
	}

	for i, def := range config.Definitions {
		if def.PriorityVal < 1 {
			def.PriorityVal = i
		}

		if def.SettingsFileVal != "" {
			def.settings = readSettingsFromTOMLFile(def.SettingsFileVal)
		} else {
			def.settings = &defaultPluginSettings{
				settings: make(map[string]interface{}),
			}
		}

		def.settings.overrideFromEnv = config.OverrideSettingsFromEnv

		def.isActive = true
	}

	sort.Sort(byPriority(config.Definitions))
	return &config, nil
}

func (dc *defaultConfig) PluginDefinitions() []ouretl.PluginDefinition {
	var definitions []ouretl.PluginDefinition
	for _, def := range dc.Definitions {
		definitions = append(definitions, def)
	}

	return definitions
}

func (dc *defaultConfig) AppendPluginDefinition(pdef ouretl.PluginDefinition) error {
	dc.Definitions = append(dc.Definitions, &defaultPluginDefinition{
		NameVal:     pdef.Name(),
		PathVal:     pdef.FilePath(),
		VersionVal:  pdef.Version(),
		PriorityVal: pdef.Priority(),
		isActive:    pdef.IsActive(),
	})

	sort.Sort(byPriority(dc.Definitions))

	for _, listener := range dc.onAddChangeListeners {
		listener(pdef)
	}

	return nil
}

func (dc *defaultConfig) OnPluginDefinitionAdded(fn func(ouretl.PluginDefinition)) {
	dc.onAddChangeListeners = append(dc.onAddChangeListeners, fn)
}

func (dc *defaultConfig) OnPluginDefinitionActivated(fn func(ouretl.PluginDefinition)) {
	dc.onActivateChangeListeners = append(dc.onActivateChangeListeners, fn)
}

func (dc *defaultConfig) OnPluginDefinitionDeactivated(fn func(ouretl.PluginDefinition)) {
	dc.onDeactivateChangeListeners = append(dc.onDeactivateChangeListeners, fn)
}

func (dc *defaultConfig) updateStatusTo(isActive bool, pdef ouretl.PluginDefinition) {
	for _, p := range dc.Definitions {
		if p.Name() == pdef.Name() && p.Version() == pdef.Version() {
			p.isActive = isActive
		}
	}
}

func (dc *defaultConfig) Activate(pdef ouretl.PluginDefinition) {
	dc.updateStatusTo(true, pdef)
	for _, listener := range dc.onActivateChangeListeners {
		listener(pdef)
	}
}

func (dc *defaultConfig) Deactivate(pdef ouretl.PluginDefinition) {
	dc.updateStatusTo(false, pdef)
	for _, listener := range dc.onDeactivateChangeListeners {
		listener(pdef)
	}
}

func (dc *defaultConfig) findAddedDefinitions(nextConfig *defaultConfig) []ouretl.PluginDefinition {
	var pdefs []ouretl.PluginDefinition

	for _, pdef := range nextConfig.PluginDefinitions() {
		currentStatus := getPluginDefinitionStatus(dc, pdef)
		nextStatus := getPluginDefinitionStatus(nextConfig, pdef)

		if nextStatus == pluginDefinitionActive && currentStatus == pluginDefinitionMissing {
			pdefs = append(pdefs, pdef)
		}
	}

	return pdefs
}

func (dc *defaultConfig) findActivatedDefinitions(nextConfig *defaultConfig) []ouretl.PluginDefinition {
	var pdefs []ouretl.PluginDefinition

	for _, pdef := range nextConfig.PluginDefinitions() {
		currentStatus := getPluginDefinitionStatus(dc, pdef)
		nextStatus := getPluginDefinitionStatus(nextConfig, pdef)

		if nextStatus == pluginDefinitionActive && currentStatus != pluginDefinitionActive {
			pdefs = append(pdefs, pdef)
		}
	}

	return pdefs
}

func (dc *defaultConfig) findRemovedDefinitions(nextConfig *defaultConfig) []ouretl.PluginDefinition {
	var pdefs []ouretl.PluginDefinition

	for _, pdef := range dc.PluginDefinitions() {
		currentStatus := getPluginDefinitionStatus(dc, pdef)
		nextStatus := getPluginDefinitionStatus(nextConfig, pdef)

		if nextStatus == pluginDefinitionMissing && currentStatus == pluginDefinitionActive {
			pdefs = append(pdefs, pdef)
		}
	}

	return pdefs
}

func (dc *defaultConfig) createFileWatch(configFilePath string) {
	w := watcher.New()
	w.FilterOps(watcher.Write)

	go func() {
		for {
			select {
			case <-w.Event:
				nextConfig, err := readConfigFromFile(configFilePath)
				if err == nil {
					added := dc.findAddedDefinitions(nextConfig)
					for _, a := range added {
						dc.AppendPluginDefinition(a)
					}

					activated := dc.findActivatedDefinitions(nextConfig)
					for _, a := range activated {
						dc.Activate(a)
					}

					removed := dc.findRemovedDefinitions(nextConfig)
					for _, r := range removed {
						dc.Deactivate(r)
					}
				}
			case err := <-w.Error:
				log.Warn(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.Add(configFilePath); err != nil {
		log.Error(err)
	}

	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Error(err)
	}
}

func getPluginDefinitionStatus(config *defaultConfig, pdef ouretl.PluginDefinition) pluginDefinitionStatus {
	for _, p := range config.PluginDefinitions() {
		if p.Name() == pdef.Name() && p.Version() == pdef.Version() && p.IsActive() {
			return pluginDefinitionActive
		}
		if p.Name() == pdef.Name() && p.Version() == pdef.Version() && !p.IsActive() {
			return pluginDefinitionInactive
		}
	}

	return pluginDefinitionMissing
}
