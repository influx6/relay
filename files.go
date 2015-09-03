package relay

// func serveFile(dir, file string, c *Context) error {
// 	fs := http.Dir(dir)
// 	f, err := fs.Open(file)
// 	if err != nil {
// 		return NewHTTPError(http.StatusNotFound)
// 	}
//
// 	fi, _ := f.Stat()
// 	if fi.IsDir() {
// 		file = spath.Join(file, indexFile)
// 		f, err = fs.Open(file)
// 		if err != nil {
// 			return NewHTTPError(http.StatusForbidden)
// 		}
// 		fi, _ = f.Stat()
// 	}
//
// 	http.ServeContent(c.response, c.request, fi.Name(), fi.ModTime(), f)
// 	return nil
// }
