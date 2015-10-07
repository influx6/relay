package cli

import (
	"fmt"
	"go/build"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

var binBuilder = func(pwd string, config *BuildConfig) error {
	pkg, err := build.Import(config.Package, "", 0)

	if err != nil {
		fmt.Printf("--> --> PkgBuild:Error: for %s -> %s with build.Import func \n", config.Package, err)
		return err
	}

	_, binName := filepath.Split(pkg.ImportPath)

	if config.Goget {
		fmt.Printf("--> Running go get for %s\n", pkg.ImportPath)
		_, err = GoDeps(pkg.ImportPath)

		if err != nil {
			fmt.Printf("--> --> go.get.Error: %s\n", err)
			return err
		}
	}

	fmt.Printf("--> Building binary file into %s\n", config.Bin)

	_, err = Gobuild(config.Bin, binName)

	if err != nil {
		fmt.Printf("--> --> go.build.Error: %s\n", err)
		return err
	}

	return nil
}

var assetsBuilder = func(pwd string, config *BuildConfig) error {
	fmt.Printf("--> Building client code in %s as %s\n", config.Client.Dir, config.ClientPackage)

	var verbose bool

	if bo, err := strconv.ParseBool(config.Client.Verbose); err == nil {
		verbose = bo
	}

	session := NewJSSession(config.VFS, config.Client.BuildTags, verbose, false)

	js, jsmap, err := session.BuildPkg(config.ClientPackage, config.Client.Name)

	if err != nil {
		fmt.Printf("--> --> go.goperjs.build.Error: for %s -> %s\n", config.ClientPackage, err)
		return err
	}

	fmt.Printf("--> Making static/js directory if not exisitng \n")

	jsdir := filepath.Join(pwd, config.Client.StaticDir)
	_ = os.MkdirAll(jsdir, 0777)

	jsfilepath := filepath.Join(jsdir, fmt.Sprintf("%s.js", config.Client.Name))
	jsmapfilepath := filepath.Join(jsdir, fmt.Sprintf("%s.js.map", config.Client.Name))

	jsfile, err := os.Create(jsfilepath)

	if err != nil {
		fmt.Printf("--> --> go.mkdir.js.file: for %s -> %s\n", jsfilepath, err)
		return err
	}

	jsfile.Write(js.Data())
	// fmt.Fprint(jsfile, js.Data())
	jsfile.Close()

	jsmapfile, err := os.Create(jsmapfilepath)

	if err != nil {
		fmt.Printf("--> --> go.mkdir.js.file: for %s -> %s\n", jsfilepath, err)
		return err
	}

	jsmapfile.Write(jsmap.Data())
	// fmt.Fprint(jsmapfile, jsmap.Data())
	jsmapfile.Close()

	fmt.Printf("--> Building static files in %s to %s/vfs_static.go \n", config.Static.Dir, config.VFS)

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

	fmt.Printf("--> Using StripPrefix (%s) for all virtual paths \n", strip)

	var done = make(chan bool)

	err = BundleStatic(virtualfile, "vfs", config.Static.Exclude, pwd, virtualsets, nil, func() error {
		close(done)
		return nil
	})

	if err != nil {
		fmt.Printf("--> --> go.mkdir.vfs.static: for %s -> %s\n", virtualfile, err)
		return err
	}

	<-done
	return nil
}

var builder = func(config *BuildConfig) error {
	pwd, _ := os.Getwd()

	if err := assetsBuilder(pwd, config); err != nil {
		fmt.Printf("--> --> cmd.assetBuilder.Error: for %s -> %s\n", config.ClientPackage, err)
		return err
	}

	if err := binBuilder(pwd, config); err != nil {
		fmt.Printf("--> --> cmd.binBuilder.Error: for %s -> %s\n", config.ClientPackage, err)
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
		pwd, _ := os.Getwd()

		fmt.Printf("--> Searching for 'app.yaml' file in (%s)....\n", pwd)

		//get the app.file
		appfile := filepath.Join(pwd, "./app.yml")

		if _, err := os.Stat(appfile); err != nil {
			fmt.Printf("--> --> Error: 'app.yml' not found in (%s)....\n", pwd)
			return
		}

		var config = NewBuildConfig()
		var ustrip = assets.EnsureSlash(filepath.Join(pwd, config.Static.StripPrefix))
		var ccop = assets.EnsureSlash(filepath.Join(pwd, config.Client.Dir))
		var acop = assets.EnsureSlash(filepath.Join(pwd, config.Static.Dir))
		var tcop = assets.EnsureSlash(filepath.Join(pwd, "templates"))
		var mainfile = assets.EnsureSlash(filepath.Join(pwd, config.Main))

		fmt.Printf("--> Found app.yml and loading into config...\n")

		if err := config.Load(appfile); err != nil {
			fmt.Printf("--> --> ConfigError: 'app.yaml' -> %s\n", err)
			return
		}

		config.Watcher.Skip = append(config.Watcher.Skip, "./vfs", "./vfs/vfs_static.go")

		// lets build the files
		if err := builder(config); err != nil {
			return
		}

		_, binName := filepath.Split(config.Package)
		bin := filepath.Join(pwd, config.Bin)

		fmt.Printf("--> Initializing binary from %s with name %s\n", bin, binName)

		var waiter sync.WaitGroup
		var cmdwaiter sync.WaitGroup
		var hasCmds bool

		//initalize it so we dont run into a blocked channel
		var runChan chan bool = make(chan bool)
		var cmdChan chan bool = make(chan bool)

		if len(config.Commands) > 0 {
			hasCmds = true
			cmdChan = RunCMD(config.Commands, func() {
				fmt.Printf("=====================CMDS RUNNED========================================\n")
				cmdwaiter.Done()
			})

			// fmt.Printf("=====================COMMANDS INITD========================================\n")
			cmdwaiter.Add(1)
			cmdChan <- true
		}

		if config.GoMain {
			fmt.Printf("--> Running Mainfile file: %s\n", mainfile)
			runChan = RunGo(mainfile, config.BinArgs, nil)
		} else {
			fmt.Printf("--> Running binary file: %s\n", filepath.Join(bin, binName))
			runChan = RunBin(bin, binName, config.BinArgs, nil)
		}

		fmt.Printf("--> Sending Run Signal to GoChan <- true \n")
		runChan <- true

		config.Watcher.Pkgs = append(config.Watcher.Pkgs, config.Package, "github.com/influx6/relay/relay", "github.com/influx6/relay/engine")

		fmt.Printf("--> Registering Packages for watching in -> (%s) \n", config.Watcher.Dir)

		dirpath := filepath.Join(pwd, config.Watcher.Dir)
		binpath := filepath.Join(pwd, config.Bin)
		vfspath := filepath.Join(pwd, config.VFS)

		var waiting int64
		var binwaiting int64

		//this watch will watch only for assets and static file changes including js changes
		assetsWatcher, err := assets.NewWatch(assets.WatcherConfig{
			Dir:      dirpath,
			Ext:      config.Watcher.Ext,
			Skip:     config.Watcher.Skip,
			MaxRetry: config.Watcher.MaxRetry,
			ExtraPkg: config.Watcher.Pkgs,
			Filter: func(addPath, basePath string) bool {
				addPath = assets.EnsureSlash(addPath)

				if strings.Index(addPath, ".git") != -1 {
					return false
				}

				if strings.Index(addPath, ustrip) == -1 {
					return false
				}

				if strings.Index(addPath, ccop) != -1 {
					return true
				}

				if strings.Index(addPath, acop) != -1 {
					return true
				}

				if strings.Index(addPath, tcop) != -1 {
					return true
				}

				// ext := filepath.Ext(addPath)
				// if ext == ".go" {
				// }

				return false
			},
		}, func(err error, ev *fsnotify.Event, wo *assets.Watcher) {
			//if its an error dont act
			if err != nil {
				return
			}

			//if we are alredy waiting just ignore this, use this instead of calling waiter.Wait()
			// because that means we will do double compile, better
			// TODO: is double compile when done after one another good?
			if atomic.LoadInt64(&waiting) > 0 {
				return
			}

			atomic.StoreInt64(&waiting, 1)
			{

				if hasCmds {
					// fmt.Printf("=====================COMMANDS INITD========================================\n")
					cmdwaiter.Add(1)
					cmdChan <- true
					cmdwaiter.Wait()
					// fmt.Printf("=====================CMD RUNNED========================================\n")
				}

				waiter.Add(1)
				fmt.Printf("=====================ASSETS========================================\n")
				fmt.Printf("File change at ->: %s, Commencing Asset Recompilation\n", ev.Name)
				fmt.Printf("=====================ASSETS========================================\n")
				fmt.Printf("\n")
				fmt.Printf("\n")

				//execute assetBuilder to build assets alone
				assetsBuilder(pwd, config)
				fmt.Printf("=====================ASSETS REBUILT========================================\n")

				waiter.Done()
			}
			atomic.StoreInt64(&waiting, 0)
		})

		binWatcher, err := assets.NewWatch(assets.WatcherConfig{
			Dir:      dirpath,
			Ext:      config.Watcher.Ext,
			Skip:     config.Watcher.Skip,
			MaxRetry: config.Watcher.MaxRetry,
			ExtraPkg: config.Watcher.Pkgs,
			Filter: func(addPath, basePath string) bool {
				addPath = assets.EnsureSlash(addPath)
				stat, err := os.Stat(addPath)
				if err == nil && stat.IsDir() {
					return true
				}

				if filepath.Ext(addPath) == config.Static.TemplateExtension {
					return true
				}

				//
				// if filepath.Ext(addPath) == ".css" {
				// 	return true
				// }

				if strings.Index(addPath, ".git") != -1 {
					return false
				}

				if !(filepath.Ext(addPath) == ".go") {
					return false
				}

				return true
			},
		}, func(err error, ev *fsnotify.Event, wo *assets.Watcher) {

			//if its an error dont act
			if err != nil {
				return
			}

			if atomic.LoadInt64(&binwaiting) > 0 {
				return
			}

			if strings.Contains(ev.Name, binpath) {
				// fmt.Printf("Skipping bin dir changes @ %s\n", ev.Name)
				return
			}

			if strings.Contains(ev.Name, vfspath) {
				// fmt.Printf("Skipping bin dir changes @ %s\n", ev.Name)
				return
			}

			atomic.StoreInt64(&binwaiting, 1)
			{
				// readyforChange = false
				fmt.Printf("=====================GO-CODE========================================\n")
				fmt.Printf("File change at ->: %s, Commencing Binary Recompilation\n", ev.Name)
				fmt.Printf("=====================GO-CODE========================================\n")
				fmt.Printf("\n")
				fmt.Printf("\n")

				waiter.Wait()

				//execute binBuilder to build binary alone
				err = binBuilder(pwd, config)
				fmt.Printf("=====================BINARY REBUILT========================================\n")

				//goroutine and relunch and restart watcher
				go func() {
					runChan <- true
					wo.Start()
				}()

			}
			atomic.StoreInt64(&binwaiting, 0)
		})

		if err != nil {
			fmt.Printf("ConfigError: loading fileWatcher -> %s\n", err)
			return
		}

		// fmt.Printf("------------------------------------done----------------------\n")
		// fmt.Printf("\n")

		go assetsWatcher.Start()
		go binWatcher.Start()

		config.Goget = false

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGQUIT)
		signal.Notify(ch, syscall.SIGTERM)
		signal.Notify(ch, os.Interrupt)

		//setup a for loop and begin calling
		for {
			select {
			case <-ch:
				fmt.Printf("Received Kill Signal, will stop watcher and application\n")
				assetsWatcher.Stop()
				binWatcher.Stop()
				close(runChan)
				close(cmdChan)
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
