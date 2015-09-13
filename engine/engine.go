package engine

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imdario/mergo"
	"github.com/influx6/flux"
	"github.com/influx6/relay"
	"gopkg.in/yaml.v2"
)

// DefaultConfig provides a default configuration for the app
var DefaultConfig = Config{
	Addr:      ":8080",
	Env:       "dev",
	Folders:   Folders{},
	Templates: relay.TemplateConfig{"./templates", nil, ".tmpl"},
	Heartbeat: "5m",
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

// UnmarshalYAML unmarshalls the incoming data for use
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
	Name      string               `yaml:"name"`
	APIToken  string               `yaml:"api_token"`
	Addr      string               `yaml:"addr"`
	Env       string               `yaml:"env"`
	Heartbeat string               `yaml:heartbeat`
	C         TLSConfig            `yaml:"tls"`
	Folders   Folders              `yaml:"folders"`
	Templates relay.TemplateConfig `yaml:"templates"`
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
		log.Printf("Unable to ReadConfig File: %s -> %s", file, err.Error())
		return err
	}

	conf := Config{}
	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		log.Printf("Unable to load Config File: %s -> %s", file, err.Error())
		return err
	}

	return mergo.MergeWithOverwrite(c, conf)
}

//Engine provides a base luncher for a service
type Engine struct {
	*relay.Routes
	*Config
	li       net.Listener
	Template *relay.TemplateDir
	//BeforeInit is run right before the server is started
	BeforeInit func(*Engine)
	//AfterInit is run right after the server is started
	AfterInit func(*Engine)
	//OnInit is runned immediate the server gets started
	OnInit func(*Engine)
	//HeartBeats is run a constant rate every ms provided
	HeartBeats func(*Engine)
}

//NewEngine returns a new app configuration
func NewEngine(c *Config, init func(*Engine)) *Engine {
	eo := &Engine{
		Config:   c,
		Routes:   relay.NewRoutes(""),
		Template: relay.NewTemplateDir(&c.Templates),
	}

	if init != nil {
		init(eo)
	}
	return eo
}

func (a *Engine) loadup() error {
	//is the asset folder not empty?, if so load it up
	if a.Folders.Assets != "" {

		log.Printf("Setting up assets dir: %s", a.Folders.Assets)
		if _, err := os.Stat(a.Folders.Assets); err != nil {
			return err
		}

		a.ServeDir("/assets", a.Folders.Assets, "/assets/")
		log.Printf("Done loading assets dir: %s", a.Folders.Assets)
	}

	if a.OnInit != nil {
		a.OnInit(a)
	}

	return nil
}

// Serve serves the app and configuration and loads the routes and serivices settings
func (a *Engine) Serve() error {
	var err error
	var li net.Listener

	//run the before init function
	if a.BeforeInit != nil {
		a.BeforeInit(a)
	}

	if a.C.Certs != nil {
		_, li, err = relay.CreateTLS(a.Addr, a.C.Certs, a)
	} else {
		_, li, err = relay.CreateHTTP(a.Addr, a)
	}

	if err != nil {
		log.Fatalf("Server failed to start: %+s", err.Error())
		return err
	}

	a.li = li

	//load up configurations
	if err := a.loadup(); err != nil {
		return err
	}

	if a.AfterInit != nil {
		a.AfterInit(a)
	}

	return nil
}

// EngineAddr returns the address of the app
func (a *Engine) EngineAddr() net.Addr {
	return a.li.Addr()
}

// Close closes and returns an error of the internal listener
func (a *Engine) Close() error {
	return a.li.Close()
}

// AppSignalInit provides a wrap function thats starts up the server and loads up,awaiting for a signal to kill
func AppSignalInit(app *Engine) {

	//start up the app server calling the .Serve()
	go flux.RecoveryHandlerCallback("App.Engine.Serve", app.Serve, func(ex interface{}) {
		//if we are in dev mode then panic,we should know when things go wrong
		log.Printf("Error occured: %s will panic if in dev env", ex)
		if app.Env == "dev" {
			panic(ex)
		}
	})

	//setup the signal block and listen for the interrup
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGQUIT)
	signal.Notify(ch, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt)

	hearbeat := makeDuration(app.Heartbeat, (10 * 60))
	//setup a for loop and begin calling
	for {
		select {
		case <-time.After(hearbeat):
			//TODO: make app return info,health status and
			//useful info
			if app.HeartBeats != nil {
				app.HeartBeats(app)
			}
		case <-ch:
			app.Close()
			os.Exit(0)
		}
	}
}
