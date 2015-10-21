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
	".haml":     "text/haml",
	".markdown": "text/markdown",
	".md":       "text/markdown",
	".svg":      "image/svg+xml",
	".png":      "image/png",
	".jpg":      "image/jpg",
	".gif":      "image/png",
	".mp3":      "audio/mp3",
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

	cext := mime.TypeByExtension(ext)

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
