package core

import (
	"io/ioutil"
	"testing"

	"github.com/ourstudio-se/ouretl-abstractions"
)

func TestThatReadConfigOfMissingFileFails(t *testing.T) {
	_, err := NewDefaultConfigFromTOMLFile("/tmp/missing-file")
	if err == nil {
		t.Errorf("Missing file did not cause error when reading config")
	}
}

func TestThatReadConfigOfTOMLFileSucceeds(t *testing.T) {
	configFilePath := "/tmp/config1.conf"
	configString := "default_plugin_path = \"/tmp/plugins\"\n\n"
	ioutil.WriteFile(configFilePath, []byte(configString), 0600)

	_, err := NewDefaultConfigFromTOMLFile(configFilePath)
	if err != nil {
		t.Error(err)
	}
}

func TestThatReadConfigPopulateValues(t *testing.T) {
	configFilePath := "/tmp/config2.conf"
	configString := "default_plugin_path = \"/tmp/plugins\"\n\n[[plugin]]\nname = \"test-1\"\npath = \"/tmp/test-1\"\nversion = \"1.0.0\"\nactive = false\n\n[[plugin]]\nname = \"test-2\"\npath = \"/tmp/test-2\"\nversion = \"1.0.0\"\nactive = false\n\n"
	ioutil.WriteFile(configFilePath, []byte(configString), 0600)

	config, err := NewDefaultConfigFromTOMLFile(configFilePath)
	if err != nil {
		t.Error(err)
	}

	if len(config.PluginDefinitions()) != 2 {
		t.Errorf("Expected plugin count of 2 did not match read value count of %d", len(config.PluginDefinitions()))
	}
	if config.PluginDefinitions()[0].Name() != "test-1" {
		t.Errorf("Expected plugin name '%s' did not match read value '%s'", "test-1", config.PluginDefinitions()[0].Name())
	}
}

func TestThatAddPluginAppendsToList(t *testing.T) {
	config := newDefaultConfig()
	_ = config.AppendPluginDefinition(&defaultPluginDefinition{})

	if len(config.PluginDefinitions()) == 0 {
		t.Errorf("Could not add a plugin definition during config runtime")
	}
}

func TestThatAppendPluginCallsChangeListener(t *testing.T) {
	config := newDefaultConfig()

	called := false
	count := 0
	config.OnPluginDefinitionAdded(func(_ ouretl.PluginDefinition) {
		count = len(config.PluginDefinitions())
		called = true
	})

	_ = config.AppendPluginDefinition(&defaultPluginDefinition{})

	if !called {
		t.Errorf("Change listener not called when adding plugin definition")
	}
	if count == 0 {
		t.Errorf("Changed config instance does not contain added plugin definition")
	}
}
