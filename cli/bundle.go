/* Bundle Drafted from (github.com/mjibson/esc)
esc
===

A file embedder for Go.

Godoc with examples: http://godoc.org/github.com/mjibson/esc
*/

package cli

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type _escFile struct {
	data     []byte
	local    string
	fileinfo os.FileInfo
}

// LoadFiles loads files from sets of giving directory lists
func LoadFiles(filesets []string, prefix, ignore string) ([]string, []string, map[string]_escFile, error) {
	var err error
	var fnames, dirnames []string
	var content = make(map[string]_escFile)

	var ignoreRegexp *regexp.Regexp

	if ignore != "" {
		ignoreRegexp, err = regexp.Compile(ignore)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	for _, base := range filesets {
		files := []string{base}
		for len(files) > 0 {
			fname := files[0]
			files = files[1:]
			if ignoreRegexp != nil && ignoreRegexp.MatchString(fname) {
				continue
			}
			f, err := os.Open(fname)
			if err != nil {
				return nil, nil, nil, err
			}
			fi, err := f.Stat()
			if err != nil {
				return nil, nil, nil, err
			}
			if fi.IsDir() {
				fis, err := f.Readdir(0)
				if err != nil {
					return nil, nil, nil, err
				}
				for _, fi := range fis {
					files = append(files, filepath.Join(fname, fi.Name()))
				}
			} else {
				b, err := ioutil.ReadAll(f)
				if err != nil {
					return nil, nil, nil, err
				}
				fpath := filepath.ToSlash(fname)
				n := strings.TrimPrefix(fpath, prefix)
				n = path.Join("/", n)
				content[n] = _escFile{data: b, local: fpath, fileinfo: fi}
				fnames = append(fnames, n)
			}
			f.Close()

		}

	}

	sort.Strings(fnames)

	return fnames, dirnames, content, nil
}

// BundleStatic generates a static file contain the set content to be bundle into a go file
func BundleStatic(Outfile, Pkg, ignore, prefix string, fileSets []string, extras []*Vfile, fx func() error) error {
	if Outfile == "" {
		panic("outfile name can not be empty")
	}
	var err error
	var fnames, dirnames []string
	var content map[string]_escFile

	prefix = filepath.ToSlash(prefix)
	fnames, dirnames, content, err = LoadFiles(fileSets, prefix, ignore)

	if err != nil {
		return err
	}

	w := bytes.NewBuffer(nil)
	// w := os.Stdout

	fmt.Fprintf(w, header, Pkg)

	dirs := map[string]bool{"/": true}

	for _, fname := range fnames {

		f := content[fname]

		for b := path.Dir(fname); b != "/"; b = path.Dir(b) {
			dirs[b] = true
		}

		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)

		if _, err := gw.Write(f.data); err != nil {
			return err
		}

		if err := gw.Close(); err != nil {
			return err
		}

		t := f.fileinfo.ModTime().Unix()

		fmt.Fprintf(w, fileform, fname, f.local, len(f.data), t, segment(&buf), "\n")
	}

	//loadup the extra forms also
	for _, vf := range extras {
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)

		if _, err := gw.Write(vf.Data()); err != nil {
			return err
		}

		if err := gw.Close(); err != nil {
			return err
		}

		fmt.Fprintf(w, fileform, vf.Name(), vf.Path(), vf.ModTime().Unix(), segment(&buf), "\n")
	}

	for d := range dirs {
		dirnames = append(dirnames, d)
	}

	sort.Strings(dirnames)

	for _, dir := range dirnames {
		local := path.Join(prefix, dir)
		if len(local) == 0 {
			local = "."
		}
		fmt.Fprintf(w, dirform, dir, local, "\n")
	}

	fmt.Fprint(w, footer)

	if Outfile != "" {
		outwriter, err := os.Create(Outfile)

		if err != nil {
			w.Reset()
			return err
		}

		defer outwriter.Close()
		io.Copy(outwriter, w)

		if fx != nil {
			return fx()
		}
		return nil
	}

	if fx != nil {
		return fx()
	}

	return nil
}

func segment(s *bytes.Buffer) string {
	var b bytes.Buffer
	b64 := base64.NewEncoder(base64.StdEncoding, &b)
	b64.Write(s.Bytes())
	b64.Close()
	res := "`\n"
	chunk := make([]byte, 80)
	for n, _ := b.Read(chunk); n > 0; n, _ = b.Read(chunk) {
		res += string(chunk[0:n]) + "\n"
	}
	return res + "`"
}
