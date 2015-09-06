package relay

import (
	"fmt"
	"net/http"
	"os"
	path "path/filepath"
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
		file = path.Join(file, indexFile)
		f, err = fs.Open(file)
		if err != nil {
			return NewCustomError("http.ServeFile.Status", fmt.Sprintf("%d", http.StatusForbidden))
		}
		fi, _ = f.Stat()
	}

	http.ServeContent(res, req, fi.Name(), fi.ModTime(), f)
	return nil
}

//AssetMap provides a map of []byte that contain already loaded asset files
type AssetMap map[string]*os.File

// AssetLoader is a function type which returns the content in []byte of a specific asset
type AssetLoader func(string) ([]byte, error)

// LoadTemplateAsset allows loading a template using a function that returns an asset
func LoadTemplateAsset(ext, delims []string, fx AssetLoader) (*template.Template, error) {

	return nil, nil
}

// LoadTemplates returns a template object with all the cached templates. You pass the extension in used (tls),the dir we need to cache and a function map for the templates
func LoadTemplates(dir string, exts, delims []string, mo template.FuncMap) (*template.Template, error) {
	return nil, nil
}
