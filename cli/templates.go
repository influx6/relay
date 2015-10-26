package cli

const (
	fileform = `
	%q: {
		local:   %q,
		size:    %v,
		modtime: %v,
		compressed: %s,
	},%s
  `

	dirform = `
	%q: {
		isDir: true,
		local: %q,
	},%s`

	appgofile = `package main

import (

	"net/http"
	"github.com/influx6/relay/relay"
	"github.com/influx6/relay/engine"
)

func main() {

	server := engine.NewEngine(engine.NewConfig(), func(app *engine.Engine) {

		app.ServeDir("/static/*", app.Static.Dir, app.Static.StripPrefix)

	})

	if err := server.Load("./app.yml"); err != nil {
		panic(err)
	}

	server.Serve()

}

`

	clientfile = `package main

import (
	"github.com/gopherjs/gopherjs/js"
)

//let this contain your js bootup script
//this will be compiled automatically and loaded with the assets files under the /vfs/client.js file

func main() {

  js.Global.Get("alert").Call("","Wecome ClientApp!")

}

`
	appYaml = `name: %s
addr: %s
env: dev

hearbeat: %dm

# directory and settings for static code
static:
  dir: ./static

# directory to locate client gopherjs code
client:
    dir: ./client

# change this to fit appropriately if using a different scheme
package: %s

# change as you see fit, this main file will be used if usemain is set to true
main: ./main.go

# enable this to use 'go run' instead of running the binary built on each rebuilding session
usemain: false
`
)

var (
	files = []string{
		"main.go",
		"client/client.go",
		"controllers/controllers.go",
	}

	directories = []string{
		"bin",
		"client",
		"client/app",
		"controllers",
		"models",
		"shared",
		"templates",
		"static",
		"vendor",
	}
)
