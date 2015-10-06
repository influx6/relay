package cli

import (
	"fmt"
	"go/build"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/influx6/assets"
	"github.com/spf13/cobra"
	"gopkg.in/fsnotify.v1"
)

var name string
var owner string

// CreateCommand creates a new relay project
var createCommand = &cobra.Command{
	Use:   "create [name] [owner]",
	Short: "creates the relay project files and directory with the given name",
	Long: `create command will generate the current directory and files with the given names and necessary configuration files

     --->{{ProjectName}}
        |--->bin
        |--->client
          |---> app
          |---> client.go
        |--->controllers
	        |--->controllers.go
        |--->templates
        |--->models
        |--->static
        |--->vfs
	        |--->vfs.go
        |--->app.yaml
        |--->main.go


    Where things like:
      client folder: contains the client code which will be generated on the fly if allowed

      static folder: contains the static files which will be generated into a ./vfs/static.go files for instant embedding
  `,
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			fmt.Printf("Error: --name flag cannot be empty '' \n")
			return
		}

		if owner == "" {
			fmt.Printf("Error: --owner flag can not be empty '' \n")
			return
		}

		fmt.Printf("-> New relay Project: %s, Owner: %s ...\n", name, owner)

		cwd, _ := os.Getwd()

		project := filepath.Join(cwd, name)

		if _, err := os.Stat(project); err == nil {
			fmt.Printf("Error: Project Folder '%s' with name already exists at %s, choose something else\n", name, project)
			return
		}

		_, reldir := filepath.Split(cwd)
		fmt.Printf("--> Creating '%s' project directory -> ./%s\n", name, reldir)

		//create project directory
		err := os.Mkdir(project, 0777)

		if err != nil {
			panic(err)
		}

		//create sub-directories from directories list
		for _, dir := range directories {
			dirpath := filepath.Join(project, dir)
			fmt.Printf("--> Creating '%s' project directory\n", dir)
			err := os.Mkdir(dirpath, 0777)
			if err != nil {
				panic(err)
			}
		}

		fmt.Printf("--> Creating project file: main.go \n")
		appmain := filepath.Join(project, "main.go")

		appmainfile, err := os.Create(appmain)

		if err != nil {
			panic(err)
		}

		maincontent := strings.Replace(appgofile, "{{controllerpkg}}", fmt.Sprintf("github.com/%s/%s/controllers", owner, name), -1)

		maincontent = strings.Replace(maincontent, "{{staticpkg}}", fmt.Sprintf("github.com/%s/%s/vfs", owner, name), -1)

		fmt.Fprintf(appmainfile, maincontent)
		appmainfile.Close()

		fmt.Printf("--> Creating project file: controllers/controllers.go \n")

		cfs := filepath.Join(project, "controllers/controllers.go")

		cfsfile, err := os.Create(cfs)

		if err != nil {
			panic(err)
		}

		fmt.Fprint(cfsfile, "package controllers")
		cfsfile.Close()

		fmt.Printf("--> Creating project file: vfs/vfs_static.go \n")

		vfs := filepath.Join(project, "vfs/vfs_static.go")

		vfsfile, err := os.Create(vfs)

		if err != nil {
			panic(err)
		}

		fmt.Fprint(vfsfile, vsgofile)
		vfsfile.Close()

		fmt.Printf("--> Creating project file: app.yml\n")
		appyml := filepath.Join(project, "app.yml")

		appfile, err := os.Create(appyml)

		if err != nil {
			panic(err)
		}

		fmt.Fprintf(appfile, appYaml, name, ":4000", 5, fmt.Sprintf("github.com/%s/%s", owner, name))
		appfile.Close()

		//lets create client/main.go file
		clientapp := filepath.Join(project, "client/client.go")

		fmt.Printf("--> Creating project file: client/client.go \n")
		cgofile, err := os.Create(clientapp)

		if err != nil {
			panic(err)
		}

		clientcontent := strings.Replace(clientfile, "{{staticpkg}}", fmt.Sprintf("github.com/%s/%s/vfs", owner, name), -1)

		clientcontent = strings.Replace(clientcontent, "{{clientpkg}}", fmt.Sprintf("github.com/%s/%s/%s/%s", owner, name, "client", "app"), -1)

		fmt.Fprint(cgofile, clientcontent)
		cgofile.Close()

		clientbase := filepath.Join(project, "client/app/app.go")

		fmt.Printf("--> Creating project file: client/app/app.go \n")

		cbofile, err := os.Create(clientbase)

		if err != nil {
			panic(err)
		}

		fmt.Fprint(cbofile, `package app`)
		cbofile.Close()

	},
}

var builder = func(config *BuildConfig) error {
	pwd, _ := os.Getwd()
	pkg, err := build.Import(config.Package, "", 0)

	if err != nil {
		fmt.Printf("PkgBuild:Error: for %s -> %s with build.Import func \n", config.Package, err)
		return err
	}

	_, binName := filepath.Split(pkg.ImportPath)

	if config.Goget {
		fmt.Printf("Running go get for %s\n", pkg.ImportPath)
		_, err = GoDeps(pkg.ImportPath)

		if err != nil {
			fmt.Printf("go.get.Error: %s\n", err)
			return err
		}
	}

	fmt.Printf("Building client code in %s as %s\n", config.Client.Dir, config.ClientPackage)

	var verbose bool

	if bo, err := strconv.ParseBool(config.Client.Verbose); err == nil {
		verbose = bo
	}

	session := NewJSSession(config.VFS, config.Client.BuildTags, verbose, false)

	js, jsmap, err := session.BuildPkg(config.ClientPackage, config.Client.Name)

	if err != nil {
		fmt.Printf("go.goperjs.build.Error: for %s -> %s\n", config.ClientPackage, err)
		return err
	}

	fmt.Printf("Making static/js directory if not exisitng \n")

	jsdir := filepath.Join(pwd, config.Client.StaticDir)
	_ = os.MkdirAll(jsdir, 0777)

	jsfilepath := filepath.Join(jsdir, fmt.Sprintf("%s.js", config.Client.Name))
	jsmapfilepath := filepath.Join(jsdir, fmt.Sprintf("%s.js.map", config.Client.Name))

	jsfile, err := os.Create(jsfilepath)

	if err != nil {
		fmt.Printf("go.mkdir.js.file: for %s -> %s\n", jsfilepath, err)
		return err
	}

	jsfile.Write(js.Data())
	// fmt.Fprint(jsfile, js.Data())
	jsfile.Close()

	jsmapfile, err := os.Create(jsmapfilepath)

	if err != nil {
		fmt.Printf("go.mkdir.js.file: for %s -> %s\n", jsfilepath, err)
		return err
	}

	jsmapfile.Write(jsmap.Data())
	// fmt.Fprint(jsmapfile, jsmap.Data())
	jsmapfile.Close()

	fmt.Printf("Building static files in %s to %s/vfs_static.go \n", config.Static.Dir, config.VFS)

	virtualfile := filepath.Join(config.VFS, "vfs_static.go")

	virtualsets := []string{
		filepath.Join(pwd, config.Static.Dir),
		filepath.Join(pwd, "templates"),
	}

	var strip string
	var ustrip = filepath.Join(pwd, config.Static.StripPrefix)

	if _, err := os.Stat(ustrip); err == nil {
		strip = ustrip
	} else {
		strip = pwd
	}

	fmt.Printf("Using StripPrefix (%s) for all virtual paths \n", strip)

	err = BundleStatic(virtualfile, "vfs", config.Static.Exclude, pwd, virtualsets, nil)

	if err != nil {
		fmt.Printf("go.mkdir.vfs.static: for %s -> %s\n", virtualfile, err)
		return err
	}

	fmt.Printf("Building binary file into %s\n", config.Bin)

	_, err = Gobuild(config.Bin, binName)

	if err != nil {
		fmt.Printf("go.build.Error: %s\n", err)
		return err
	}

	return nil
}

// BuildCommand builds the relay project
var buildCommand = &cobra.Command{
	Use:   "build",
	Short: "build the current relay project into a binary",
	Long:  `build takes all assets and static files, compiles all js into go static package and builds a new binary that contains all this together`,
	Run: func(cmd *cobra.Command, args []string) {
		pwd, _ := os.Getwd()

		fmt.Printf("Searching for 'app.yaml' file in (%s)....\n", pwd)

		//get the app.file
		appfile := filepath.Join(pwd, "./app.yml")

		if _, err := os.Stat(appfile); err != nil {
			fmt.Printf("Error: 'app.yml' not found in (%s)....\n", pwd)
			return
		}

		var config = NewBuildConfig()

		fmt.Printf("Found app.yml and loading into config...\n")

		if err := config.Load(appfile); err != nil {
			fmt.Printf("ConfigError: 'app.yaml' -> %s\n", err)
			return
		}

		builder(config)
	},
}

// ServeCommand builds the relay project
var serveCommand = &cobra.Command{
	Use:   "serve",
	Short: "serves up the project and watches for changes",
	Long:  `it will rebuild and bundle your project files with build and reserve them on any change`,
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		var wa *assets.Watcher

		pwd, _ := os.Getwd()

		fmt.Printf("Searching for 'app.yaml' file in (%s)....\n", pwd)

		//get the app.file
		appfile := filepath.Join(pwd, "./app.yml")

		if _, err := os.Stat(appfile); err != nil {
			fmt.Printf("Error: 'app.yml' not found in (%s)....\n", pwd)
			return
		}

		var config = NewBuildConfig()

		fmt.Printf("Found app.yml and loading into config...\n")

		if err := config.Load(appfile); err != nil {
			fmt.Printf("ConfigError: 'app.yaml' -> %s\n", err)
			return
		}

		config.Watcher.Skip = append(config.Watcher.Skip, "./vfs", "./vfs/static.go")

		fmt.Printf("................Build............................\n")

		//lets build the files
		err = builder(config)

		if err != nil {
			fmt.Printf("PkgBuild:Error: for %s -> %s with build.Import func \n", config.Package, err)
			return
		}

		_, binName := filepath.Split(config.Package)
		bin := filepath.Join(pwd, config.Bin)

		fmt.Printf("Initializing binary from %s with name %s\n", bin, binName)

		fmt.Printf("Running binary file: %s\n", filepath.Join(bin, binName))
		runChan := RunBin(bin, binName, config.BinArgs, nil)

		fmt.Printf("Sending Run Signal to GoChan <- true \n")
		runChan <- true

		config.Watcher.Pkgs = append(config.Watcher.Pkgs, config.Package, "github.com/influx6/relay/relay", "github.com/influx6/relay/engine")
		fmt.Printf("Generating assets for : %s -> Packages: %+s\n", config.Watcher.Dir, config.Watcher.Pkgs)

		dirpath := filepath.Join(pwd, config.Watcher.Dir)
		binpath := filepath.Join(pwd, config.Bin)

		var readyforChange = false

		wa, err = assets.NewWatch(assets.WatcherConfig{
			Dir:      dirpath,
			Ext:      config.Watcher.Ext,
			Skip:     config.Watcher.Skip,
			MaxRetry: config.Watcher.MaxRetry,
			ExtraPkg: config.Watcher.Pkgs,
		}, func(err error, ev *fsnotify.Event, wo *assets.Watcher) {

			//if its an error stop
			if err != nil {
				return
			}

			if !readyforChange {
				return
			}

			if strings.Contains(ev.Name, binpath) {
				// fmt.Printf("Skipping bin dir changes @ %s\n", ev.Name)
				return
			}

			readyforChange = false
			fmt.Printf("File change at ->: %s, Commencing Recompilation\n", ev.Name)

			fmt.Printf("\n.............Build: COMPILING CHANGE...............................\n\n")
			//execute build command
			err = builder(config)

			//goroutine and relunch and restart watcher
			go func() {
				runChan <- true
				fmt.Printf("............................................\n\n")
				readyforChange = true
				wo.Start()
			}()

		})

		if err != nil {
			fmt.Printf("ConfigError: loading fileWatcher -> %s\n", err)
			return
		}

		fmt.Printf("............................................\n\n")
		go wa.Start()
		readyforChange = true

		config.Goget = false

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGQUIT)
		signal.Notify(ch, syscall.SIGTERM)
		signal.Notify(ch, os.Interrupt)

		//setup a for loop and begin calling
		for {
			select {
			case <-ch:
				fmt.Printf("Received Kill Signal, will stop watcher and app\n")
				wa.Stop()
				close(runChan)
				return
			}
		}

	},
}

// RootCmd provides the core command for the cli
var RootCmd = &cobra.Command{
	Use:   "relay",
	Short: "relay provides a cli for relay projects",
}

func init() {
	//assin the create flags
	createCommand.Flags().StringVar(&owner, "owner", "", "owner of the project used in construct the addr: github.com/owner/projectName")
	createCommand.Flags().StringVar(&name, "name", "", "name for the project")

	//add the build command to the server
	// serveCommand.AddCommand(buildCommand)
	//loadup all commands to the root command
	RootCmd.AddCommand(buildCommand, serveCommand, createCommand)
}
