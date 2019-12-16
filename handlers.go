package core

import (
	"plugin"
	"time"

	ouretl "github.com/ourstudio-se/ouretl-abstractions"
	log "github.com/sirupsen/logrus"
)

type wrapper struct {
	definition     ouretl.PluginDefinition
	implementation ouretl.DataHandlerPlugin
}

func NewHandler(definition ouretl.PluginDefinition, config ouretl.Config) *wrapper {
	p, err := plugin.Open(definition.FilePath())
	if err != nil {
		log.Errorf("Plugin '%s (v%s)' could not be found at path '%s'", definition.Name(), definition.Version(), definition.FilePath())
		return nil
	}

	actor, err := p.Lookup("GetHandler")
	if err != nil {
		log.Debugf("Plugin '%s (v%s)' did not expose a `GetHandler` symbol -- it will be excluded from messaging pipeline", definition.Name(), definition.Version())
		return nil
	}

	retriever, ok := actor.(func(ouretl.Config, ouretl.PluginSettings) (ouretl.DataHandlerPlugin, error))
	if !ok {
		log.Errorf("Plugin '%s (v%s)' was loaded as a `DataHandlerPlugin`, but does not expose a valid function declaration -- it will be excluded from messaging pipeline", definition.Name(), definition.Version())
		return nil
	}

	handler, err := retriever(config, definition.Settings())
	if err != nil {
		log.Errorf("Plugin '%s (v%s)' could not be loaded as a `DataHandlerPlugin`, received error: %v", definition.Name(), definition.Version(), err)
		return nil
	}

	log.Infof("Plugin '%s (v%s)' successfully loaded as a `DataHandlerPlugin`", definition.Name(), definition.Version())

	return &wrapper{
		definition:     definition,
		implementation: handler,
	}
}

func NewHandlerPool(config ouretl.Config) []*wrapper {
	var pool []*wrapper
	for _, definition := range config.PluginDefinitions() {
		wrapper := NewHandler(definition, config)
		if wrapper == nil {
			continue
		}

		pool = append(pool, wrapper)
	}

	return pool
}

func NewHandlerPoolFromConfig(channel <-chan *DefaultDataMessage, config ouretl.Config) {
	pool := NewHandlerPool(config)

	config.OnPluginDefinitionAdded(func(pdef ouretl.PluginDefinition) {
		wrapper := NewHandler(pdef, config)
		if wrapper != nil {
			pool = append(pool, wrapper)
			log.Infof("`DataHandlerPlugin` '%s (v%s)' added, a total of %d `DataHandlerPlugin` implementations loaded", pdef.Name(), pdef.Version(), len(pool))
		}
	})

	log.Infof("%d `DataHandlerPlugin` implementations loaded", len(pool))

	for {
		select {
		case msg := <-channel:
			proxyDataMessage(pool, msg)
		}
	}
}

func proxyDataMessage(pool []*wrapper, dm *DefaultDataMessage) {
	log.Debugf("Processing a new message with ID '%s', initiated from worker '%s'", dm.ID(), dm.Origin())
	startedAt := time.Now()

	counter := 0
	caller := func(data []byte) error {
		ms := int64(time.Since(startedAt) / time.Millisecond)
		log.Debugf("Message with ID '%s' processed by %d DataHandlerPlugin implementations in %d ms", dm.ID(), counter, ms)

		return nil
	}
	for i := (len(pool) - 1); i >= 0; i-- {
		if !pool[i].definition.IsActive() {
			log.Debugf("`DataHandlerPlugin` '%s (v%s)' is marked as INACTIVE", pool[i].definition.Name(), pool[i].definition.Version())
			continue
		}

		counter = counter + 1
		next := newDataFunc(pool[i], dm, caller)
		caller = next
	}

	err := caller(dm.Data())
	if err != nil {
		log.Error(err)
	}
}

func newDataFunc(w *wrapper, dm *DefaultDataMessage, fn func([]byte) error) func(data []byte) error {
	return func(data []byte) error {
		log.Debugf("DataHandlerPlugin '%s (v%s)' receiving message with ID '%s'", w.definition.Name(), w.definition.Version(), dm.ID())
		return w.implementation.Handle(dm.withData(data), fn)
	}
}

func containsHandler(haystack []*wrapper, needle string) bool {
	for _, x := range haystack {
		if x.definition.Name() == needle {
			return true
		}
	}

	return false
}
