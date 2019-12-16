package core

import (
	"sync"
	"testing"

	ouretl "github.com/ourstudio-se/ouretl-abstractions"
)

type mockPluginDef struct {
	active bool
}

func (m *mockPluginDef) Name() string {
	return "plugin"
}

func (m *mockPluginDef) FilePath() string {
	return "/tmp/plugin.so.1.0.0"
}

func (m *mockPluginDef) Version() string {
	return "1.0.0"
}

func (m *mockPluginDef) Priority() int {
	return 1
}

func (m *mockPluginDef) IsActive() bool {
	return m.active
}

func (m *mockPluginDef) Settings() ouretl.PluginSettings {
	return nil
}

type mockPluginImpl struct {
	called  bool
	handled func()
}

func (m *mockPluginImpl) Handle(dm ouretl.DataMessage, next func([]byte) error) error {
	m.called = true
	m.handled()
	return next(dm.Data())
}

func TestThatPushedMessageReachesActiveHandlersInPipeline(t *testing.T) {
	config := newDefaultConfig()
	pool := newHandlerPool(config)

	var wg sync.WaitGroup
	wg.Add(3)

	p1i := &mockPluginImpl{called: false, handled: func() {
		wg.Done()
	}}
	p1 := &wrapper{
		definition:     &mockPluginDef{active: true},
		implementation: p1i,
	}
	p2i := &mockPluginImpl{called: false, handled: func() {
		wg.Done()
	}}
	p2 := &wrapper{
		definition:     &mockPluginDef{active: true},
		implementation: p2i,
	}
	p3i := &mockPluginImpl{called: false, handled: func() {
		wg.Done()
	}}
	p3 := &wrapper{
		definition:     &mockPluginDef{active: true},
		implementation: p3i,
	}

	pool = append(pool, p1)
	pool = append(pool, p2)
	pool = append(pool, p3)
	proxyDataMessage(pool, &defaultDataMessage{id: "test", data: []byte("test")})

	wg.Wait()

	if !p1i.called {
		t.Error("First data handler wasn't called")
	}
	if !p2i.called {
		t.Error("Second data handler wasn't called")
	}
	if !p3i.called {
		t.Error("Third data handler wasn't called")
	}
}

func TestThatPushedMessageDoesntReachInactiveHandlersInPipeline(t *testing.T) {
	config := newDefaultConfig()
	pool := newHandlerPool(config)

	var wg sync.WaitGroup
	wg.Add(2)

	p1i := &mockPluginImpl{called: false, handled: func() {
		wg.Done()
	}}
	p1 := &wrapper{
		definition:     &mockPluginDef{active: true},
		implementation: p1i,
	}
	p2i := &mockPluginImpl{called: false, handled: func() {
		wg.Done()
	}}
	p2 := &wrapper{
		definition:     &mockPluginDef{active: false},
		implementation: p2i,
	}
	p3i := &mockPluginImpl{called: false, handled: func() {
		wg.Done()
	}}
	p3 := &wrapper{
		definition:     &mockPluginDef{active: true},
		implementation: p3i,
	}

	pool = append(pool, p1)
	pool = append(pool, p2)
	pool = append(pool, p3)
	proxyDataMessage(pool, &defaultDataMessage{id: "test", data: []byte("test")})

	wg.Wait()

	if !p1i.called {
		t.Error("First data handler wasn't called")
	}
	if p2i.called {
		t.Error("Second data handler was called despite being inactive")
	}
	if !p3i.called {
		t.Error("Third data handler wasn't called")
	}
}
