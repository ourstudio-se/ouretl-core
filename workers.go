package core

import (
	"plugin"
	"time"

	ouretl "github.com/ourstudio-se/ouretl-abstractions"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
)

func NewWorker(definition ouretl.PluginDefinition, config ouretl.Config) ouretl.WorkerPlugin {
	p, err := plugin.Open(definition.FilePath())
	if err != nil {
		log.Errorf("Plugin '%s (v%s)' could not be found at path '%s': %s, definition.Name(), definition.Version(), definition.FilePath(), err.Error()")
		return nil
	}

	actor, err := p.Lookup("GetWorker")
	if err != nil {
		log.Debugf("Plugin '%s (v%s)' did not expose a `GetWorker` symbol -- it will be excluded from worker pool", definition.Name(), definition.Version())
		return nil
	}

	retriever, ok := actor.(func(ouretl.Config, ouretl.PluginSettings) (ouretl.WorkerPlugin, error))
	if !ok {
		log.Errorf("Plugin '%s (v%s)' was loaded as a `WorkerPlugin`, but does not expose a valid function declaration -- it will be excluded from worker pool", definition.Name(), definition.Version())
		return nil
	}

	worker, err := retriever(config, definition.Settings())
	if err != nil {
		log.Errorf("Plugin '%s (v%s)' could not be loaded as a `WorkerPlugin`, received error: %v", definition.Name(), definition.Version(), err)
		return nil
	}

	log.Infof("Plugin '%s (v%s)' successfully loaded as a `WorkerPlugin`", definition.Name(), definition.Version())

	return worker
}

func NewWorkerPool(channel chan<- *DefaultDataMessage, config ouretl.Config) []string {
	var sources []string
	for _, definition := range config.PluginDefinitions() {
		worker := NewWorker(definition, config)
		if worker == nil {
			continue
		}

		sources = append(sources, definition.Name())
		startWorker(worker, channel, definition.Name())
	}

	return sources
}

func NewWorkerPoolFromConfig(channel chan<- *DefaultDataMessage, config ouretl.Config) {
	pool := NewWorkerPool(channel, config)

	config.OnPluginDefinitionAdded(func(pdef ouretl.PluginDefinition) {
		worker := NewWorker(pdef, config)
		if worker != nil {
			pool = append(pool, pdef.Name())
			startWorker(worker, channel, pdef.Name())
			log.Infof("`DataHandlerPlugin` '%s (v%s)' added, a total of %d `DataHandlerPlugin` implementations loaded", pdef.Name(), pdef.Version(), len(pool))
		}
	})

	log.Infof("%d `WorkerPlugin` implementations loaded", len(pool))
}

func startWorker(worker ouretl.WorkerPlugin, channel chan<- *DefaultDataMessage, name string) {
	proxy := newMessageProxy(channel, name)
	go initiateWorker(worker, proxy, name)
}

func initiateWorker(worker ouretl.WorkerPlugin, proxy func([]byte), name string) {
	err := worker.Start(proxy)
	if err != nil {
		log.Error(err)
		log.Infof("Restarting worker '%s'...", name)

		time.Sleep(1 * time.Second)
		initiateWorker(worker, proxy, name)
	} else {
		log.Warnf("WorkerPlugin '%s' has exited without error", name)
	}
}

func newMessageProxy(channel chan<- *DefaultDataMessage, name string) func([]byte) {
	return func(data []byte) {
		dataMessage := &DefaultDataMessage{
			id:     uuid.NewV4().String(),
			data:   data,
			origin: name,
		}
		channel <- dataMessage
	}
}

func containsWorker(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}

	return false
}
