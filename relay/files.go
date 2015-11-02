package relay

import (
	"fmt"
	"mime"
	"net/http"
	filepath "path/filepath"
	"strings"
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
