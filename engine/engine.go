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
	Addr:    ":8080",
	Folders: Folders{},
}

//TLSConfig provides a base config for tls configuration
type TLSConfig struct {
	Certs *tls.Config
	Key   string `yaml:"key"`
	Cert  string `yaml:"cert"`
}

type tlsconf struct {
	Key  string `yaml:"key"`
	Cert string `yaml:"cert"`
}

// UnmarshalYaml unmarshalls the incoming data for use
func (t *TLSConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	toc := tlsconf{}

	if err := unmarshal(&toc); err != nil {
		return err
	}

	if toc.Cert == "" || toc.Key == "" {
		return nil
	}

	co, err := relay.LoadTLS(toc.Cert, toc.Key)

	if err != nil {
		return err
	}

	t.Certs = co
	return nil
}

// Folders provide a configuration for app-used folders
type Folders struct {
	Assets string `yaml:"assets"`
	Models string `yaml:"models"`
	Views  string `yaml:"views"`
}

// Config provides configuration for Afro
type Config struct {
	Addr    string    `yaml:"addr"`
	C       TLSConfig `yaml:"tls"`
	Folders Folders   `yaml:"folders"`
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

//Engine provides a base luncher for a service
type Engine struct {
	*relay.Routes
	li     net.Listener
	config *Config
}

//NewEngine returns a new app configuration
func NewEngine(c *Config) *Engine {
	return &Engine{
		Routes: relay.NewRoutes(""),
		config: c,
	}
}

// Serve serves the app and configuration and loads the routes and serivices settings
func (a *Engine) Serve() error {
	var err error
	var li net.Listener

	if a.config.C.Certs != nil {
		_, li, err = relay.CreateTLS(a.config.Addr, a.config.C.Certs, a)
	} else {
		_, li, err = relay.CreateHTTP(a.config.Addr, a)
	}

	if err != nil {
		log.Fatalf("Server failed to start: %+s", err.Error())
		return err
	}

	a.li = li

	return nil
}

// Addr returns the address of the app
func (a *Engine) Addr() net.Addr {
	return a.li.Addr()
}

// Close closes and returns an error of the internal listener
func (a *Engine) Close() error {
	return a.li.Close()
}
