package relay

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

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

	var tree = make(AssetMap)

	//do we have a directory?
	if !stat.IsDir() {

		var fext string
		var rel = filepath.Base(dir)

		if strings.Index(rel, ".") != -1 {
			fext = filepath.Ext(rel)
		}

		if fext == ext {
			tree[rel] = filepath.ToSlash(dir)
		}

	} else {
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

	}

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
