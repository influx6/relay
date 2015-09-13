package relay

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
)

// TemplateConfig provides a nice configuration for TemplateDir
type TemplateConfig struct {
	Dir        string
	Delimeters []string
	Extension  string
}

// DefaultTemplateConfig provides a default TemplateConfig
var DefaultTemplateConfig = TemplateConfig{
	Dir:       "./templates",
	Extension: ".tml",
}

// NewTemplateConfig returns a new TemplateConfig instance
func NewTemplateConfig() *TemplateConfig {
	t := DefaultTemplateConfig
	return &t
}

// LoadJSON loads the configuration from a yaml file
func (t *TemplateConfig) LoadJSON(file string) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Printf("Unable to ReadConfig File: %s -> %s", file, err.Error())
		return err
	}

	conf := TemplateConfig{}
	err = json.Unmarshal(data, &conf)

	if err != nil {
		log.Printf("Unable to load Config File: %s -> %s", file, err.Error())
		return err
	}

	return mergo.MergeWithOverwrite(t, conf)
}

// LoadYAML loads the configuration from a yaml file
func (t *TemplateConfig) LoadYAML(file string) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Printf("Unable to ReadConfig File: %s -> %s", file, err.Error())
		return err
	}

	conf := TemplateConfig{}
	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		log.Printf("Unable to load Config File: %s -> %s", file, err.Error())
		return err
	}

	return mergo.MergeWithOverwrite(t, conf)
}

// TemplateDir provides a struct for managing template initialization per directory.
//it allows you to set a root directory to initialize templates from
type TemplateDir struct {
	dir    string
	config *TemplateConfig
}

// NewTemplateDir returns a new TemplateDir instance
func NewTemplateDir(c *TemplateConfig) *TemplateDir {
	return &TemplateDir{filepath.Clean(c.Dir), c}
}

// Create returns a new AssetTemplate with the name and path set
func (t *TemplateDir) Create(name string, paths []string, fo []template.FuncMap) (*AssetTemplate, error) {
	var dirs []string

	for _, ps := range paths {
		dirs = append(dirs, filepath.Join(t.dir, ps))
	}

	bo := BuildAssetTemplate(name, t.config.Extension, dirs, fo, t.config.Delimeters)
	return bo, bo.Build()
}

// CreateExt returns a new AssetTemplate with the name and path set but allowing setting the custom extension
func (t *TemplateDir) CreateExt(name, ext string, paths []string, fo []template.FuncMap) (*AssetTemplate, error) {
	var dirs []string

	for _, ps := range paths {
		dirs = append(dirs, filepath.Join(t.dir, ps))
	}

	bo := BuildAssetTemplate(name, ext, dirs, fo, t.config.Delimeters)

	return bo, bo.Build()
}

// AssetTemplate provides a simple means of loading template assets and providing reloading systems which simplifies the use
type AssetTemplate struct {
	loaded bool
	name   string
	files  []string
	ext    string
	delim  []string
	amaps  []AssetMap
	Tmpl   *template.Template
	Funcs  []template.FuncMap
}

// NewAssetTemplate returns a new asset template
func NewAssetTemplate(name, ext string, dirs []string) (*AssetTemplate, error) {
	as := BuildAssetTemplate(name, ext, dirs, nil, nil)
	return as, as.Build()
}

// BuildAssetTemplate loads up a new AssetTemplate with the necessarily settings and builds its up.It allows creating a section combination of layout and included template assets
func BuildAssetTemplate(name, ext string, tpaths []string, fo []template.FuncMap, delim []string) *AssetTemplate {
	return &AssetTemplate{
		files: tpaths,
		ext:   ext,
		delim: delim,
	}
}

// Build loads up the trees if not loaded then builds up a new template with the layouts and includes if it fails returns an error
func (a *AssetTemplate) Build() error {
	if !a.loaded {
		a.amaps = a.amaps[:0]

		for _, dir := range a.files {

			include, err := AssetTree(dir, a.ext)

			if err != nil {
				return err
			}

			a.amaps = append(a.amaps, include)
		}

		a.loaded = true
	}

	tl, err := LoadTemplateAsset(a.name, a.delim, a.amaps, a.Funcs)

	if err != nil {
		return err
	}

	a.Tmpl = tl
	return nil
}

//Reload reloads the files from the given directory paths
func (a *AssetTemplate) Reload() {
	a.loaded = false
}

// Watch creates a watcher for those files for change and reloads the template files
// func (a *AssetTemplate) Watch() {
//
// }

// LoadTemplateAsset allows loading a template using a function that returns an asset
func LoadTemplateAsset(name string, delims []string, mxa []AssetMap, fx []template.FuncMap) (*template.Template, error) {

	// log.Printf("template Dir: %+s", dir)
	var tree = template.New(name)

	//check if the delimiter array has content if so,set them
	if len(delims) > 0 && len(delims) >= 2 {
		tree.Delims(delims[0], delims[1])
	}

	for _, mx := range mxa {
		// func(mx AssetMap){
		for nm := range mx {
			// func(name string,file string)

			content, err := mx.Load(nm)

			if err != nil {
				return nil, err
			}

			// log.Printf("Loading: %s", nm)

			//we dont want to throw a panic so instead lets cache it and then return a error instead
			var panicd bool
			var msg interface{}

			func(name string) {

				// log.Printf("adding: %s", name)
				tl := tree.New(name)

				for _, fu := range fx {
					func(fur template.FuncMap) {
						tl.Funcs(fur)
					}(fu)
				}

				func() {
					defer func() {
						if err := recover(); err != nil {
							msg = err
							panicd = true
						}
					}()
					template.Must(tl.Parse(string(content)))
				}()

			}(nm)

			if panicd {
				return nil, NewCustomError("Template.Must.Panic", fmt.Sprintf("Template paniced with %+s", msg))
			}
			// }(n,f)
		}
		// }(mxg)
	}

	return tree, nil
}

// LoadTemplates returns a template object with all the cached templates. You pass the extension in used (tls),the dir we need to cache and a function map for the templates
func LoadTemplates(dir, ext string, delims []string, mo []template.FuncMap) (*template.Template, error) {
	am, err := AssetTree(dir, ext)
	if err != nil {
		return nil, err
	}
	return LoadTemplateAsset(dir, delims, []AssetMap{am}, mo)
}
