package cli

import (
	"errors"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

// DefaultBuilder provides a default config for builders
var DefaultBuilder = BuildConfig{
	Addr: ":8000",
	Env:  "dev",
	// VFS:     "./vfs",
	Bin:     "./bin",
	Main:    "./main.go",
	DoGoGet: "true",
	UseMain: "false",
	Client: JSConfig{
		Dir:       "./client",
		StaticDir: "static/js",
		Name:      "client",
		Verbose:   "false",
	},
	Static: StaticConfig{
		Dir:               "./static",
		TemplateExtension: ".tmpl",
	},
	Watcher: WatcherConfig{
		Dir:      "./",
		MaxRetry: 5,
	},
}

// WatcherConfig to be used with watcher
type WatcherConfig struct {
	Dir      string   `yaml:"-"`
	Ext      []string `yaml:"ext"`
	Skip     []string `yaml:"skip"`
	Pkgs     []string `yaml:"pkgs"`
	Max      string   `yaml:"max"`
	MaxRetry int      `yaml:"-"`
	Retry    int      `yaml:"-"`
}

// JSConfig provides the configuration details for the gopherjs project location and arguments
type JSConfig struct {
	Dir        string   `yaml:"dir"`
	Exclude    []string `yaml:"exclude"`
	Extensions []string `yaml:"exts"`
	//BuildTags to be used in building the js files
	BuildTags []string `yaml:"tags"`
	//Name defines the name to call the file and its map i.e  'appa' makes js file to be appa.js and appa.js.map respectively
	Name string `yaml:"name"`
	// StaticDir defines the location in the static directory the compiled version will also be stored for cases when you decide to not use the virtual fs
	StaticDir string `yaml:"static_dir"`
	//Verbose sets the verbosity of the build process
	Verbose string `yaml:"verbose"`
}

// StaticConfig provides the configuration details for the static files location and arguments
type StaticConfig struct {
	Dir               string   `yaml:"dir"`
	StripPrefix       string   `yaml:"strip_prefix"`
	Exclude           string   `yaml:"exclude"`
	Extensions        []string `yaml:"exts"`
	TemplateExtension string   `yaml:"template_ext"`
}

// BuildConfig provides the configuration details for the building constraints for using relay's builder
type BuildConfig struct {
	Name string `yaml:"name"`
	Addr string `yaml:"addr"`
	Env  string `yaml:"env"`
	// VFS           string        `yaml:"vfs"`
	Bin           string        `yaml:"bin"`
	Main          string        `yaml:"main"`
	Package       string        `yaml:"package"`
	BinArgs       []string      `yaml:"bin_args"`
	Client        JSConfig      `yaml:"client"`
	Static        StaticConfig  `yaml:"static"`
	Watcher       WatcherConfig `yaml:"watcher"`
	ClientPackage string        `yaml:"-"`
	Goget         bool          `yaml:"-"`
	GoMain        bool          `yaml:"-"`
	//Commands will be executed before any building of assets or compiling of binary
	Commands []string `yaml:"commands"`
	//dogoget ensures that after the first initial building that go get gets re-run on each rebuild
	DoGoGet string `yaml:"dogoget"`
	//useMain ensures to instead run the main file giving in 'main' instead of the built binary to reduce time
	UseMain string `yaml:"usemain"`
}

// NewBuildConfig returns a new BuildConfig based off the defaults
func NewBuildConfig() *BuildConfig {
	bc := DefaultBuilder
	return &bc
}

// Load loads the configuration from a yaml file
func (c *BuildConfig) Load(file string) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Printf("Unable to ReadConfig File: %s -> %s", file, err.Error())
		return err
	}

	conf := BuildConfig{}
	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		log.Printf("Unable to load Config File: %s -> %s", file, err.Error())
		return err
	}

	if conf.Package == "" {
		return errors.New("package option can not be empty, provide the project package name please")
	}

	if err := mergo.MergeWithOverwrite(c, conf); err != nil {
		return err
	}

	if mano, err := strconv.ParseBool(c.UseMain); err == nil {
		c.GoMain = mano
	}

	if doge, err := strconv.ParseBool(c.DoGoGet); err == nil {
		c.Goget = doge
	}

	if max, err := strconv.Atoi(c.Watcher.Max); err == nil {
		c.Watcher.MaxRetry = max
	}

	c.ClientPackage = filepath.Join(c.Package, "client")

	if strings.Contains(c.ClientPackage, `\`) {
		c.ClientPackage = filepath.ToSlash(c.ClientPackage)
	}

	return nil
}
