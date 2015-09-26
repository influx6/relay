package engine

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imdario/mergo"
	"github.com/influx6/assets"
	"github.com/influx6/flux"
	"github.com/influx6/relay"
	"github.com/tylerb/graceful"
	"gopkg.in/yaml.v2"
)

// DefaultConfig provides a default configuration for the app
var DefaultConfig = Config{
	Addr:      ":8080",
	Env:       "dev",
	Folders:   Folders{},
	Heartbeat: "5m",
	Killbeat:  "2m",
	Templates: assets.TemplateConfig{
		Dir:       "./templates",
		Extension: ".tmpl",
	},
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

// Db provides a generic db configuration value
type Db struct {
	Type string `yaml:"type"` //can be 'sql','mgo
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
	User string `yaml:"user"`
	Pass string `yaml:"pass"`
}

// Config provides configuration for Afro
type Config struct {
	Name      string `yaml:"name"`
	APIToken  string `yaml:"api_token"`
	Addr      string `yaml:"addr"`
	Env       string `yaml:"env"`
	Heartbeat string `yaml:"heartbeat"`
	//the timeout for graceful shutdown of server
	Killbeat  string                `yaml:"killbeat"`
	C         TLSConfig             `yaml:"tls"`
	Folders   Folders               `yaml:"folders"`
	Db        Db                    `yaml:"db"`
	Templates assets.TemplateConfig `yaml:"templates"`
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
	li              *graceful.Server
	ls              net.Listener
	stop, heartbeat time.Duration
	Template        *assets.TemplateDir
	//HeartBeats is run a constant rate every ms provided
	HeartBeats func(*Engine)
	//BeforeInit is run right before the server is started
	BeforeInit func(*Engine)
	//AfterInit is run right after the server is started
	AfterInit func(*Engine)
	//OnInit is runned immediate the server gets started
	OnInit  func(*Engine)
	OnClose func(*Engine)
}

//NewEngine returns a new app configuration
func NewEngine(c *Config, init func(*Engine)) *Engine {
	eo := &Engine{
		Config:   c,
		Routes:   relay.NewRoutes(""),
		Template: assets.NewTemplateDir(&c.Templates),
		OnInit:   init,
	}

	eo.stop = makeDuration(c.Killbeat, 20)
	eo.heartbeat = makeDuration(c.Heartbeat, (10 * 60))

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
func (a *Engine) Serve() {
	//start up the app server calling the .Serve()
	if err := a.prepareServer(); err != nil {
		panic(err)
	}

	log.Printf("Application %s running @ %s", a.Name, a.Addr)

	//setup the signal block and listen for the interrup
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGQUIT)
	signal.Notify(ch, syscall.SIGTERM)
	signal.Notify(ch, os.Interrupt)

	//setup a for loop and begin calling
	for {
		select {
		case <-time.After(a.heartbeat):
			//useful info
			if a.HeartBeats != nil {
				a.HeartBeats(a)
			}
		case <-ch:
			a.Close()
			return
		}
	}
}

func (a *Engine) prepareServer() error {
	var err error
	var ls net.Listener

	//run the before init function
	if a.BeforeInit != nil {
		a.BeforeInit(a)
	}

	ls, err = relay.MakeBaseListener(a.Addr, a.C.Certs)

	if err != nil {
		log.Fatalf("Server failed to start: %+s", err.Error())
		return err
	}

	a.ls = ls
	a.li = &graceful.Server{
		NoSignalHandling: true,
		Timeout:          a.stop,
		Server: &http.Server{
			Addr:    a.Addr,
			Handler: a,
		},
	}

	flux.GoDefer("ServerGracefulServer", func() {
		a.li.Serve(ls)
	})

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
	if a.ls == nil {
		return nil
	}
	return a.ls.Addr()
}

// Close closes and returns an error of the internal listener
func (a *Engine) Close() error {
	defer func() {
		if a.OnClose != nil {
			a.OnClose(a)
		}
	}()

	if a.li == nil {
		return os.ErrInvalid
	}
	a.li.Stop(a.stop)
	return nil
}
