package relay

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	filepath "path/filepath"
	"strings"

	"github.com/influx6/reggy"
)

var mediaTypes = map[string]string{
	".txt":      "text/plain",
	".text":     "text/plain",
	".html":     "text/html",
	".css":      "text/css",
	".js":       "text/javascript",
	".erb":      "template/erb",
	".min.css":  "text/css",
	".haml":     "text/haml",
	".markdown": "text/markdown",
	".md":       "text/markdown",
	".svg":      "image/svg+xml",
	".png":      "image/png",
	".jpg":      "image/jpg",
	".gif":      "image/png",
	".mp3":      "audio/mp3",
}

// FS provides a configurable struct that provides a http.FileSystem with extra customization options
type FS struct {
	http.FileSystem
	Strip  string
	Header http.Header
}

// FSCtxHandler returns a valid FlatHandler
func FSCtxHandler(fs *FS) FlatHandler {
	return FSContextHandler(fs, nil)
}

// FSContextHandler returns a valid FlatHandler
func FSContextHandler(fs *FS, fail http.HandlerFunc) FlatHandler {
	fsh := FSHandler(fs, fail)
	return func(c *Context, next NextHandler) {
		fsh(c.Res, c.Req, c.c)
		next(c)
	}
}

// FSHandler returns a valid Route.RHandler for use with the relay.Route
func FSHandler(fs *FS, fail http.HandlerFunc) RHandler {
	return func(res http.ResponseWriter, req *http.Request, c Collector) {
		strip := path.Clean(fmt.Sprintf("/%s/", fs.Strip))
		requested := reggy.CleanPath(req.URL.Path)
		file := strings.TrimPrefix(requested, strip)

		fi, err := fs.Open(file)

		if err != nil {
			if fail != nil {
				fail(res, req)
			} else {
				http.NotFound(res, req)
			}
			return
		}

		ext := strings.ToLower(filepath.Ext(file))

		stat, err := fi.Stat()

		if err != nil {
			if fail != nil {
				fail(res, req)
			} else {
				http.NotFound(res, req)
			}
			return
		}

		// var cext string
		cext := mime.TypeByExtension(ext)

		// log.Printf("Type ext: %s -> %s", ext, cext)
		if cext == "" {
			if types, ok := mediaTypes[ext]; ok {
				cext = types
			} else {
				cext = "text/plain"
			}
		}

		res.Header().Set("Content-Type", cext)

		if fs.Header != nil {
			for m, v := range fs.Header {
				for _, va := range v {
					res.Header().Add(m, va)
				}
			}
		}

		http.ServeContent(res, req, stat.Name(), stat.ModTime(), fi)
	}
}

// FSServe provides a http.Handler for serving using a http.FileSystem
func FSServe(fs http.FileSystem, stripPrefix string, fail http.HandlerFunc) RHandler {
	if fail == nil {
		fail = http.NotFound
	}

	return func(res http.ResponseWriter, req *http.Request, c Collector) {
		strip := path.Clean(fmt.Sprintf("/%s/", stripPrefix))
		requested := reggy.CleanPath(req.URL.Path)
		file := strings.TrimPrefix(requested, strip)

		fi, err := fs.Open(file)

		if err != nil {
			fail(res, req)
			return
		}

		ext := strings.ToLower(filepath.Ext(file))

		stat, err := fi.Stat()

		if err != nil {
			fail(res, req)
			return
		}

		// var cext string
		cext := mime.TypeByExtension(ext)

		// log.Printf("Type ext: %s -> %s", ext, cext)
		if cext == "" {
			if types, ok := mediaTypes[ext]; ok {
				cext = types
			} else {
				cext = "text/plain"
			}
		}

		res.Header().Set("Content-Type", cext)
		http.ServeContent(res, req, stat.Name(), stat.ModTime(), fi)
	}
}

// UseFS returns a custom http.FileSystem with extra extensions in tailoring response
func UseFS(fs http.FileSystem, hd http.Header, strip string) *FS {
	fsm := FS{
		FileSystem: fs,
		Strip:      strip,
		Header:     hd,
	}
	return &fsm
}

// NewFS returns a custom http.FileSystem with extra extensions in tailoring response
func NewFS(fs http.FileSystem, strip string) *FS {
	fsm := FS{
		FileSystem: fs,
		Strip:      strip,
		Header:     make(http.Header),
	}
	return &fsm
}

// ServeFile provides a file handler for serving files, it takes an indexFile which defines a default file to look for if the file path is a directory ,then the directory to use and the file to be searched for
func ServeFile(indexFile, dir, file string, res http.ResponseWriter, req *http.Request) error {
	fs := http.Dir(dir)
	f, err := fs.Open(file)

	// log.Printf("opening file from %s for %s -> %s", dir, file, err)

	if err != nil {
		return NewCustomError("http.ServeFile.Status", fmt.Sprintf("%d", http.StatusNotFound))
	}

	ext := strings.ToLower(filepath.Ext(file))

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexFile)
		f, err = fs.Open(file)
		if err != nil {
			return NewCustomError("http.ServeFile.Status", fmt.Sprintf("%d", http.StatusForbidden))
		}
		fi, _ = f.Stat()
	}

	// var cext string
	cext := mime.TypeByExtension(ext)

	// log.Printf("Type ext: %s -> %s", ext, cext)
	if cext == "" {
		if types, ok := mediaTypes[ext]; ok {
			cext = types
		} else {
			cext = "text/plain"
		}
	}

	res.Header().Set("Content-Type", cext)

	http.ServeContent(res, req, fi.Name(), fi.ModTime(), f)
	return nil
}
