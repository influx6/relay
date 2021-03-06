package cli

import (
	"fmt"

	"github.com/influx6/flux"
)

// PluginMux defines a function type for a plugin activator
type PluginMux func(*BuildConfig, Plugins, chan bool)

// PluginManager provides a basic plugin management system for the cli
type PluginManager struct {
	plugins flux.Collector
}

// NewPluginManager returns a new plugin manager
func NewPluginManager() *PluginManager {
	pm := PluginManager{plugins: flux.NewCollector()}
	return &pm
}

// Add adds a new plugin to the list
func (pm *PluginManager) Add(name string, fx PluginMux) {
	pm.plugins.Set(name, fx)
}

// Activate activates a specific plugin
func (pm *PluginManager) Activate(m Plugins, b *BuildConfig, c chan bool) {
	br := pm.plugins.Get(m.Tag)
	if br != nil {
		if bx, ok := br.(PluginMux); ok {
			fmt.Printf("--> Plugin: Initializing Plugin: '%s' \n", m.Tag)
			bx(b, m, c)
		}
	}
}
