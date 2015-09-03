package engine

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"

	"github.com/imdario/mergo"
	"github.com/influx6/relay"
	"gopkg.in/yaml.v2"
)

// DefaultConfig provides a default configuration for the app
var DefaultConfig = Config{
	Addr:   ":8080",
	UseTLS: false,
	Certs:  TLSConfig{},
}

//TLSConfig provides a base config for tls configuration
type TLSConfig struct {
	conf *tls.Config
	key  string
	cert string
}

// UnmarshalYaml unmarshalls the incoming data for use
func (t *TLSConfig) UnmarshalYaml(unmarshal func(interface{}) error) error {
	// var k, c string

	// if err := unmarshal()
	return nil
}

// Config provides configuration for Afro
type Config struct {
	Addr   string
	UseTLS bool
	Certs  TLSConfig
}

// NewConfig returns a new configuration file
func NewConfig() *Config {
	c := DefaultConfig
	return &c
}

// Load loads the configuration from a yaml file
func (c *Config) Load(file string) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	conf := Config{}
	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		return err
	}

	return mergo.MergeWithOverwrite(c, conf)
}

// UnmarshalYAML unmarshals and sets the configuration options
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {

	return nil
}

//Engine provides a base luncher for a service
type Engine struct {
	*relay.Routes
	li     net.Listener
	config *Config
}

//NewEngine returns a new app configuration
func NewEngine(c *Config) *Engine {
	return &Engine{
		Routes: relay.NewRoutes(),
		config: c,
	}
}

// Serve serves the app and configuration and loads the routes and serivices settings
func (a *Engine) Serve() {
	var err error
	var li net.Listener

	if a.config.UseTLS && a.config.Certs.conf != nil {
		_, li, err = relay.CreateTLS(a.config.Addr, a.config.Certs.conf, a)
	} else {
		_, li, err = relay.CreateHTTP(a.config.Addr, a)
	}

	if err != nil {
		log.Fatalf("Server failed to start: %+s", err.Error())
		return
	}

	a.li = li
}
