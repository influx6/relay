package relay

import (
	"fmt"
	"mime"
	"net/http"
	filepath "path/filepath"
)

// ServeFile provides a file handler for serving files, it takes an indexFile which defines a default file to look for if the file path is a directory ,then the directory to use and the file to be searched for
func ServeFile(indexFile, dir, file string, res http.ResponseWriter, req *http.Request) error {
	fs := http.Dir(dir)
	f, err := fs.Open(file)

	// log.Printf("opening file from %s for %s -> %s", dir, file, err)

	if err != nil {
		return NewCustomError("http.ServeFile.Status", fmt.Sprintf("%d", http.StatusNotFound))
	}

	ext := filepath.Ext(file)

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexFile)
		f, err = fs.Open(file)
		if err != nil {
			return NewCustomError("http.ServeFile.Status", fmt.Sprintf("%d", http.StatusForbidden))
		}
		fi, _ = f.Stat()
	}

	// cext := mime.TypeByExtension(ext)
	// if cext != "" {
	res.Header().Add("Content-Type", mime.TypeByExtension(ext))
	// }
	http.ServeContent(res, req, fi.Name(), fi.ModTime(), f)
	return nil
}
