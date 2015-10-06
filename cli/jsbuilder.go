package cli

import (
	"bytes"
	"errors"
	"os"
	"path"
	"time"

	build "github.com/gopherjs/gopherjs/build"
	"github.com/gopherjs/gopherjs/compiler"
	"github.com/neelance/sourcemap"
)

// ErrNotMain is returned when we find no .go file with 'main' package
var ErrNotMain = errors.New("Package contains no 'main' go package file")

// Vfile or virtual file for provide a virtual file info
type Vfile struct {
	path string
	name string
	data []byte
	mod  time.Time
}

// Path returns the path of the file
func (v *Vfile) Path() string {
	return path.Join(v.path, v.name)
}

// Name returns the name of the file
func (v *Vfile) Name() string {
	return v.name
}

// Sys returns nil
func (v *Vfile) Sys() interface{} {
	return nil
}

// Data returns the data captured within
func (v *Vfile) Data() []byte {
	return v.data
}

// Mode returns 0 as the filemode
func (v *Vfile) Mode() os.FileMode {
	return 0
}

// Size returns the size of the data
func (v *Vfile) Size() int64 {
	return int64(len(v.data))
}

// ModTime returns the modtime for the virtual file
func (v *Vfile) ModTime() time.Time {
	return v.mod
}

// IsDir returns false
func (v *Vfile) IsDir() bool {
	return false
}

// JSSession represents a basic build.Session with its option
type JSSession struct {
	//Dir to use for the virtual files
	dir     string
	Session *build.Session
	Option  *build.Options
}

// NewJSSession returns a new session for build js files
func NewJSSession(dir string, tags []string, verbose, watch bool) *JSSession {

	options := &build.Options{
		Verbose:       verbose,
		Watch:         watch,
		CreateMapFile: true,
		Minify:        true,
		BuildTags:     tags,
	}

	session := build.NewSession(options)

	return &JSSession{
		dir:     dir,
		Session: session,
		Option:  options,
	}
}

// BuildPkg uses the session, to build a package file with the given output name and returns two virtual files containing the js and js.map respectively, or an error
func (j *JSSession) BuildPkg(pkg, name string) (*Vfile, *Vfile, error) {
	var js, jsmap *bytes.Buffer = bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	if err := BuildJS(j, pkg, name, js, jsmap); err != nil {
		return nil, nil, err
	}

	jsv, jsmapv := MakeVFiles(name, j.dir, js, jsmap)
	return jsv, jsmapv, nil
}

// BuildDir uses the session, to build a particular dir contain files and using the specified package name and output name returns two virtual files containing the js and js.map respectively, or an error
func (j *JSSession) BuildDir(dir, importpath, name string) (*Vfile, *Vfile, error) {
	var js, jsmap *bytes.Buffer = bytes.NewBuffer(nil), bytes.NewBuffer(nil)

	if err := BuildJSDir(j, dir, importpath, name, js, jsmap); err != nil {
		return nil, nil, err
	}

	jsv, jsmapv := MakeVFiles(name, j.dir, js, jsmap)
	return jsv, jsmapv, nil
}

// MakeVFiles turns js and jsmap buffer into virtual files
func MakeVFiles(name, path string, js, jsmap *bytes.Buffer) (*Vfile, *Vfile) {
	return &Vfile{
			name: name,
			path: path,
			data: js.Bytes(),
			mod:  time.Now(),
		}, &Vfile{
			name: name + ".js.map",
			path: path,
			data: jsmap.Bytes(),
			mod:  time.Now(),
		}
}

// BuildJSDir builds the js file and returns the content.
// goPkgPath must be a package path eg. github.com/influx6/haiku-examples/app
func BuildJSDir(jsession *JSSession, dir, importpath, name string, js, jsmap *bytes.Buffer) error {

	session, options := jsession.Session, jsession.Option

	buildpkg, err := build.NewBuildContext(session.InstallSuffix(), options.BuildTags).ImportDir(dir, 0)

	if err != nil {
		return err
	}

	pkg := &build.PackageData{Package: buildpkg}
	pkg.ImportPath = importpath

	//build the package using the sessios
	if err = session.BuildPackage(pkg); err != nil {
		return err
	}

	//build up the source map also
	smfilter := &compiler.SourceMapFilter{Writer: js}
	smsrc := &sourcemap.Map{File: name + ".js"}
	smfilter.MappingCallback = build.NewMappingCallback(smsrc, options.GOROOT, options.GOPATH)
	deps, err := compiler.ImportDependencies(pkg.Archive, session.ImportContext.Import)

	if err != nil {
		return err
	}

	err = compiler.WriteProgramCode(deps, smfilter)

	smsrc.WriteTo(jsmap)
	js.WriteString("//# sourceMappingURL=" + name + ".map.js\n")

	return nil
}

// BuildJS builds the js file and returns the content.
// goPkgPath must be a package path eg. github.com/influx6/haiku-examples/app
func BuildJS(jsession *JSSession, goPkgPath, name string, js, jsmap *bytes.Buffer) error {

	session, options := jsession.Session, jsession.Option

	//get the build path
	buildpkg, err := build.Import(goPkgPath, 0, session.InstallSuffix(), options.BuildTags)

	if err != nil {
		return err
	}

	if buildpkg.Name != "main" {
		return ErrNotMain
	}

	//build the package data for building
	pkg := &build.PackageData{Package: buildpkg}

	//build the package using the sessios
	if err = session.BuildPackage(pkg); err != nil {
		return err
	}

	//build up the source map also
	smfilter := &compiler.SourceMapFilter{Writer: js}
	smsrc := &sourcemap.Map{File: name + ".js"}
	smfilter.MappingCallback = build.NewMappingCallback(smsrc, options.GOROOT, options.GOPATH)
	deps, err := compiler.ImportDependencies(pkg.Archive, session.ImportContext.Import)

	if err != nil {
		return err
	}

	err = compiler.WriteProgramCode(deps, smfilter)

	smsrc.WriteTo(jsmap)
	js.WriteString("//# sourceMappingURL=" + name + ".map.js\n")

	return nil
}
