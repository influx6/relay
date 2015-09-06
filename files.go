package relay

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	filepath "path/filepath"
	"strings"
	"text/template"
)

// ServeFile provides a file handler for serving files, it takes an indexFile which defines a default file to look for if the file path is a directory ,then the directory to use and the file to be searched for
func ServeFile(indexFile, dir, file string, res http.ResponseWriter, req *http.Request) error {
	fs := http.Dir(dir)
	f, err := fs.Open(file)
	if err != nil {
		return NewCustomError("http.ServeFile.Status", fmt.Sprintf("%d", http.StatusNotFound))
	}

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexFile)
		f, err = fs.Open(file)
		if err != nil {
			return NewCustomError("http.ServeFile.Status", fmt.Sprintf("%d", http.StatusForbidden))
		}
		fi, _ = f.Stat()
	}

	http.ServeContent(res, req, fi.Name(), fi.ModTime(), f)
	return nil
}

//AssetMap provides a map of paths that contain assets of the specific filepaths
type AssetMap map[string]string

// AssetTree provides a map tree of files across the given directory that match the filenames being used
func AssetTree(dir string, ext string) (AssetMap, error) {
	var stat os.FileInfo
	var err error

	//do the path exists
	if stat, err = os.Stat(dir); err != nil {
		return nil, err
	}

	//do we have a directory?
	if !stat.IsDir() {
		return nil, NewCustomError("AssetTree", fmt.Sprintf("path is not a directory: %s", dir))
	}

	var tree = make(AssetMap)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		//if info is nil or is a directory when we skip
		if info == nil || info.IsDir() {
			return nil
		}

		var rel string
		var rerr error

		//is this path relative to the current one, if not,return err
		if rel, rerr = filepath.Rel(dir, path); rerr != nil {
			return rerr
		}

		var fext string

		if strings.Index(rel, ".") != -1 {
			fext = filepath.Ext(rel)
		}

		if fext == ext {
			tree[rel] = filepath.ToSlash(path)
			// tree[strings.TrimSuffix(rel, ext)] = filepath.ToSlash(path)
		}

		return nil
	})

	return tree, nil
}

// AssetLoader is a function type which returns the content in []byte of a specific asset
type AssetLoader func(string) ([]byte, error)

// Has returns true/false if the filename without its extension exists
func (am AssetMap) Has(name string) bool {
	_, ok := am[name]
	return ok
}

// Load returns the data of the specific file with the name
func (am AssetMap) Load(name string) ([]byte, error) {
	if !am.Has(name) {
		return nil, NewCustomError("AssetMap", fmt.Sprintf("%s unknown", name))
	}
	return ioutil.ReadFile(am[name])
}

// LoadTemplateAsset allows loading a template using a function that returns an asset
func LoadTemplateAsset(dir string, delims []string, mxa []AssetMap, fx []template.FuncMap) (*template.Template, error) {

	// log.Printf("template Dir: %+s", dir)
	var tree = template.New(dir)

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
