package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-fsnotify/fsnotify"

	"github.com/influx6/assets"
	"github.com/influx6/flux"
	"github.com/influx6/reactors/builders"
	"github.com/influx6/reactors/fs"
)

// RegisterDefaultPlugins provides a set of default plugins for relay
func RegisterDefaultPlugins(pm *PluginManager) {
	addBuilder(pm)
	addGoFriday(pm)
	addGoStaticBundle(pm)
	addJSWatchBuild(pm)
	addJsClient(pm)
	addWatchBuildRun(pm)
	addCommander(pm)
}

func addBuilder(pm *PluginManager) {
	//these are internally used
	pm.Add("builder", func(config *BuildConfig, options Plugins, c chan bool) {
		pwd, _ := os.Getwd()
		_, binName := filepath.Split(config.Package)
		// bin := filepath.Join(pwd, config.Bin)
		var clientdir string

		outputdir := filepath.Join(pwd, config.Client.StaticDir)

		if config.Client.Dir != "" {
			clientdir = filepath.Join(pwd, config.Client.Dir)
		}

		goget := builders.GoInstallerWith("./")

		jsbuild := builders.JSLauncher(builders.JSBuildConfig{
			Package:    config.ClientPackage,
			Folder:     outputdir,
			FileName:   config.Client.Name,
			Tags:       config.Client.BuildTags,
			Verbose:    config.Client.UseVerbose,
			PackageDir: clientdir,
		})

		gobuild := builders.GoBuilderWith(builders.BuildConfig{
			Path: filepath.Join(pwd, config.Bin),
			Name: binName,
			Args: config.BinArgs,
		})

		goget.Bind(jsbuild, true)

		//send out the build command after js build
		jsbuild.React(func(root flux.Reactor, _ error, _ interface{}) {
			gobuild.Send(true)
		}, true)

		//run go installer
		goget.Send(true)

		flux.GoDefer("watchBuildRun:kill", func() {
			<-c
			//close our builders
			goget.Close()
			gobuild.Close()
		})
	})
}

func addWatchBuildRun(pm *PluginManager) {
	//these are internally used
	pm.Add("watchBuildRun", func(config *BuildConfig, options Plugins, c chan bool) {
		pwd, _ := os.Getwd()
		_, binName := filepath.Split(config.Package)
		binDir := filepath.Join(pwd, config.Bin)
		binfile := filepath.Join(binDir, binName)

		pkgs := append([]string{}, config.Package, "github.com/influx6/relay/relay", "github.com/influx6/relay/engine")

		packages, err := assets.GetAllPackageLists(pkgs)

		if err != nil {
			panic(err)
		}

		fmt.Printf("--> Retrieved package directories %s \n", config.Package)

		goget := builders.GoInstallerWith("./")
		goget.React(func(root flux.Reactor, err error, data interface{}) {
			if err != nil {
				fmt.Printf("---> goget.Error occured: %s\n", err)
			} else {
				fmt.Printf("--> Sending signal for 'go get'\n")
			}
		}, true)

		buildbin := builders.BinaryBuildLauncher(builders.BinaryBuildConfig{
			Path:    binDir,
			Name:    binName,
			RunArgs: config.BinArgs,
		})

		buildbin.React(func(root flux.Reactor, err error, data interface{}) {
			if err != nil {
				fmt.Printf("---> buildbin.Error occured: %s\n", err)
			} else {
				fmt.Printf("--> Building Binary")
			}
		}, true)

		goget.Bind(buildbin, true)

		fmt.Printf("--> Initializing File Watcher using package dependecies at %d\n", len(packages))

		watcher := fs.WatchSet(fs.WatchSetConfig{
			Path: packages,
			Validator: func(base string, info os.FileInfo) bool {
				if strings.Contains(base, ".git") {
					return false
				}

				if strings.Contains(base, binDir) || base == binDir {
					return false
				}

				if strings.Contains(base, binfile) || base == binfile {
					return false
				}

				if info != nil && info.IsDir() {
					return true
				}

				if filepath.Ext(base) != ".go" {
					return false
				}

				return true
			},
		})

		watcher.React(func(root flux.Reactor, err error, data interface{}) {
			if err != nil {
				fmt.Printf("---> watcher.Error occured: %s\n", err)
			} else {
				if ev, ok := data.(fsnotify.Event); ok {
					fmt.Printf("--> File as changed: %+s\n", ev.String())
				}
			}
		}, true)

		watcher.Bind(buildbin, true)
		watcher.Bind(goget, true)

		//run go installer
		goget.Send(true)

		fmt.Printf("--> Initializing Interrupt Signal  Watcher for %s@%s\n", binName, binfile)

		flux.GoDefer("watchBuildRun:kill", func() {
			<-c
			//close our builders
			watcher.Close()
			goget.Close()
			buildbin.Close()
		})
	})
}

func addJSWatchBuild(pm *PluginManager) {
	//these are internally used for js building
	pm.Add("jsWatchBuild", func(config *BuildConfig, options Plugins, c chan bool) {
		pwd, _ := os.Getwd()
		_, binName := filepath.Split(config.Package)
		binDir := filepath.Join(pwd, config.Bin)
		binfile := filepath.Join(binDir, binName)

		pkgs := append([]string{}, config.ClientPackage)

		packages, err := assets.GetAllPackageLists(pkgs)

		if err != nil {
			panic(err)
		}

		// packages = append(packages, pwd)
		fmt.Printf("--> Retrieved js package directories %s \n", config.Package)

		var clientdir string

		outputdir := filepath.Join(pwd, config.Client.StaticDir)

		if config.Client.Dir != "" {
			clientdir = filepath.Join(pwd, config.Client.Dir)
		}

		jsbuild := builders.JSLauncher(builders.JSBuildConfig{
			Package:    config.ClientPackage,
			Folder:     outputdir,
			FileName:   config.Client.Name,
			Tags:       config.Client.BuildTags,
			Verbose:    config.Client.UseVerbose,
			PackageDir: clientdir,
		})

		jsbuild.React(func(root flux.Reactor, err error, _ interface{}) {
			if err != nil {
				fmt.Printf("--> Js.client.Build complete: Dir: %s \n -----> Error: %s \n", clientdir, err)
			}
		}, true)

		fmt.Printf("--> Initializing File Watcher using js package dependecies at %d\n", len(packages))

		watcher := fs.WatchSet(fs.WatchSetConfig{
			Path: packages,
			Validator: func(base string, info os.FileInfo) bool {
				if strings.Contains(base, ".git") {
					return false
				}

				if strings.Contains(base, binDir) || base == binDir {
					return false
				}

				if strings.Contains(base, binfile) || base == binfile {
					return false
				}

				if info != nil && info.IsDir() {
					return true
				}

				if filepath.Ext(base) != ".go" {
					return false
				}

				// log.Printf("allowed: %s", base)
				return true
			},
		})

		watcher.React(flux.SimpleMuxer(func(root flux.Reactor, data interface{}) {
			if ev, ok := data.(fsnotify.Event); ok {
				fmt.Printf("--> Client:File as changed: %+s\n", ev.String())
			}
		}), true)

		watcher.Bind(jsbuild, true)

		jsbuild.Send(true)

		flux.GoDefer("jsWatchBuild:kill", func() {
			<-c
			//close our builders
			watcher.Close()
			jsbuild.Close()
		})
	})
}

func addGoFriday(pm *PluginManager) {
	pm.Add("goFriday", func(config *BuildConfig, options Plugins, c chan bool) {
		/*Expects to receive a plugin config follow this format

		      tag: gofriday
		      config:
		        markdown: ./markdown
		        templates: ./templates

		  		  where the config.path is the path to be watched

		*/

		//get the current directory
		pwd, _ := os.Getwd()

		//get the dir we should watch
		markdownDir := options.Config["markdown"]
		templateDir := options.Config["templates"]

		//optional args
		ext := options.Config["ext"]
		//must be a bool
		sanitizeString := options.Config["sanitize"]

		var sanitize bool

		if svz, err := strconv.ParseBool(sanitizeString); err == nil {
			sanitize = svz
		}

		if markdownDir == "" || templateDir == "" {
			fmt.Println("---> gofriday.error: expected to find keys (markdown and templates) in config map")
			return
		}

		//get the absolute path
		absDir := filepath.Join(pwd, markdownDir)
		tbsDir := filepath.Join(pwd, templateDir)

		gofriday, err := builders.GoFridayStream(builders.MarkStreamConfig{
			InputDir: absDir,
			SaveDir:  tbsDir,
			Ext:      ext,
			Sanitize: sanitize,
		})

		if err != nil {
			fmt.Printf("---> gofriday.error: %s", err)
			return
		}

		//create the file watcher
		watcher := fs.Watch(fs.WatchConfig{
			Path: absDir,
		})

		watcher.React(flux.SimpleMuxer(func(root flux.Reactor, data interface{}) {
			if ev, ok := data.(fsnotify.Event); ok {
				fmt.Printf("--> goFriday:File as changed: %+s\n", ev.String())
			}
		}), true)
		// create the command runner set to run the args
		watcher.Bind(gofriday, true)

		flux.GoDefer("goFiday:kill", func() {
			<-c
			watcher.Close()
		})
	})
}

func addGoStaticBundle(pm *PluginManager) {
	pm.Add("goStatic", func(config *BuildConfig, options Plugins, c chan bool) {
		/*Expects to receive a plugin config follow this format: you can control all aspects of the assets.BindFS using the following

		      tag: gostatic
					# add commands to run on file changes
					args:
						- touch ./templates/smirf.go
		      config:
		        in: ./markdown
		        out: ./templates
						package: smirf
						file: smirf
						gzipped: true
						nodecompression: true
						production: true // generally you want to leave this to the cli to set

		  		  where the config.path is the path to be watched

		*/

		//get the current directory
		pwd, _ := os.Getwd()

		//get the dir we should watch
		inDir := options.Config["in"]
		outDir := options.Config["out"]
		packageName := options.Config["package"]
		fileName := options.Config["file"]
		ignore := options.Config["ignore"]
		absDir := filepath.Join(pwd, inDir)
		absFile := filepath.Join(pwd, outDir, fileName+".go")

		if inDir == "" || outDir == "" || packageName == "" || fileName == "" {
			fmt.Println("---> goStatic.error: the following keys(in,out,package,file) must not be empty")
			return
		}

		//set up the boolean values
		var prod bool
		var gzip bool
		var nodcom bool
		var err error

		if gz, err := strconv.ParseBool(options.Config["gzipped"]); err == nil {
			gzip = gz
		} else {
			if config.Mode > 0 {
				gzip = true
			}
		}

		if br, err := strconv.ParseBool(options.Config["nodecompression"]); err == nil {
			nodcom = br
		}

		if pr, err := strconv.ParseBool(options.Config["production"]); err == nil {
			prod = pr
		} else {
			if config.Mode <= 0 {
				prod = false
			} else {
				prod = true
			}
		}

		var ignoreReg *regexp.Regexp

		if ignore != "" {
			ignoreReg = regexp.MustCompile(ignore)
		}

		gostatic, err := builders.BundleAssets(&assets.BindFSConfig{
			InDir:           inDir,
			OutDir:          outDir,
			Package:         packageName,
			File:            fileName,
			Gzipped:         gzip,
			NoDecompression: nodcom,
			Production:      prod,
		})

		if err != nil {
			fmt.Printf("---> goStatic.error: %s", err)
			return
		}

		gostatic.React(func(root flux.Reactor, err error, data interface{}) {
			fmt.Printf("--> goStatic.Reacted: State %t Error: (%+s)\n", data, err)
		}, true)

		//bundle up the assets for the main time
		gostatic.Send(true)

		var command []string

		if prod {
			if runtime.GOOS != "windows" {
				command = append(command, fmt.Sprintf("touch %s", absFile))
			} else {
				command = append(command, fmt.Sprintf("copy /b %s+,,", absFile))
				// command = append(command, fmt.Sprintf("powershell  (ls %s).LastWriteTime = Get-Date", absFile))
			}
		}

		//add the args from the options
		command = append(command, options.Args...)
		// log.Printf("command %s", command)

		//adds a CommandLauncher to touch the output file to force a file change notification
		touchCommand := builders.CommandLauncher(command)
		gostatic.Bind(touchCommand, true)

		//create the file watcher
		watcher := fs.Watch(fs.WatchConfig{
			Path: absDir,
			Validator: func(path string, info os.FileInfo) bool {
				if ignoreReg != nil && ignoreReg.MatchString(path) {
					return false
				}
				return true
			},
		})

		// create the command runner set to run the args
		watcher.Bind(gostatic, true)

		watcher.React(flux.SimpleMuxer(func(root flux.Reactor, data interface{}) {
			if ev, ok := data.(fsnotify.Event); ok {
				fmt.Printf("--> goStatic:File as changed: %+s\n", ev.String())
			}
		}), true)

		flux.GoDefer("goStatic:kill", func() {
			<-c
			gostatic.Close()
		})
	})
}

func addCommander(pm *PluginManager) {
	pm.Add("commandWatch", func(config *BuildConfig, options Plugins, c chan bool) {
		/*Expects to receive a plugin config follow this format

		  tag: dirWatch
		  config:
		    path: "./static/less"
		  args:
		    - lessc ./static/less/main.less ./static/css/main.css
		    - lessc ./static/less/svg.less ./static/css/svg.css

		  where the config.path is the path to be watched

		*/

		//get the current directory
		pwd, _ := os.Getwd()

		//get the dir we should watch
		dir := options.Config["path"]

		//get the command we should run on change
		commands := options.Args

		if dir == "" {
			fmt.Printf("---> dirWatch.error: no path set in config map for plug")
			return
		}

		//get the absolute path
		absDir := filepath.Join(pwd, dir)

		//create the file watcher
		watcher := fs.Watch(fs.WatchConfig{
			Path: absDir,
		})

		watcher.React(flux.SimpleMuxer(func(root flux.Reactor, data interface{}) {
			if ev, ok := data.(fsnotify.Event); ok {
				fmt.Printf("--> commandWatch:File as changed: %+s\n", ev.String())
			}
		}), true)
		// create the command runner set to run the args
		watcher.Bind(builders.CommandLauncher(commands), true)

		flux.GoDefer("CommandWatch:kill", func() {
			<-c
			watcher.Close()
		})
	})
}

func addJsClient(pm *PluginManager) {
	//these are internally used for js building
	pm.Add("jsClients", func(config *BuildConfig, options Plugins, c chan bool) {
		for _, pkg := range options.Args {
			var pg Plugins
			pg.Config = make(PluginConfig)
			pg.Tag = "jsClient"
			pg.Config["package"] = pkg
			pg.Args = nil
			pm.Activate(pg, config, c)
		}
	})

	pm.Add("jsClient", func(config *BuildConfig, options Plugins, c chan bool) {
		pkg := options.Config["package"]

		_, jsName := filepath.Split(pkg)

		pkgs := append([]string{}, pkg)
		packages, err := assets.GetAllPackageLists(pkgs)

		if err != nil {
			panic(err)
		}

		dir, err := assets.GetPackageDir(pkg)

		if err != nil {
			panic(err)
		}

		jsbuild := builders.JSLauncher(builders.JSBuildConfig{
			Package:  pkg,
			Folder:   dir,
			FileName: jsName,
			Tags:     options.Args,
			Verbose:  config.Client.UseVerbose,
		})

		jsbuild.React(func(root flux.Reactor, err error, _ interface{}) {
			if err != nil {
				fmt.Printf("--> Js.client.Build complete: Dir: %s \n -----> Error: %s \n", dir, err)
			}
		}, true)

		watcher := fs.WatchSet(fs.WatchSetConfig{
			Path: packages,
			Validator: func(base string, info os.FileInfo) bool {
				if strings.Contains(base, ".git") {
					return false
				}

				if info != nil && info.IsDir() {
					return true
				}

				if filepath.Ext(base) != ".go" {
					return false
				}

				return true
			},
		})

		watcher.React(flux.SimpleMuxer(func(root flux.Reactor, data interface{}) {
			if ev, ok := data.(fsnotify.Event); ok {
				fmt.Printf("--> Client:File as changed: %+s\n", ev.String())
			}
		}), true)

		watcher.Bind(jsbuild, true)

		jsbuild.Send(true)

		flux.GoDefer("jsClient:kill", func() {
			<-c
			watcher.Close()
			jsbuild.Close()
		})
	})
}
