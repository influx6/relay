package engine

import (
	"io/ioutil"
	"log"

	"github.com/imdario/mergo"
	"github.com/influx6/assets"
	"gopkg.in/yaml.v2"
)

// DefaultClientConfig provides a default configuration for the app
var DefaultClientConfig = ClientConfig{
	Addr:      "",
	Env:       "dev",
	Folders:   Folders{},
	Templates: assets.TemplateConfig{"./templates", nil, ".tmpl"},
}

// ClientConfig provides configuration for Afro
type ClientConfig struct {
	Name      string                `yaml:"name"`
	Folders   Folders               `yaml:"folders"`
	APIToken  string                `yaml:"api_token"`
	Addr      string                `yaml:"addr"`
	Env       string                `yaml:"env"`
	Templates assets.TemplateConfig `yaml:"templates"`
}

// NewClientConfig returns a new configuration file
func NewClientConfig() *ClientConfig {
	c := DefaultClientConfig
	return &c
}

// Load loads the configuration from a yaml file
func (c *ClientConfig) Load(file string) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Printf("Unable to ReadConfig File: %s -> %s", file, err.Error())
		return err
	}

	conf := ClientConfig{}
	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		log.Printf("Unable to load Config File: %s -> %s", file, err.Error())
		return err
	}

	return mergo.MergeWithOverwrite(c, conf)
}
