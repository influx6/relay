package cli

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
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

		fmt.Fprintf(appmainfile, appgofile)
		appmainfile.Close()

		fmt.Printf("--> Creating project file: controllers/controllers.go \n")

		cfs := filepath.Join(project, "controllers/controllers.go")

		cfsfile, err := os.Create(cfs)

		if err != nil {
			panic(err)
		}

		fmt.Fprint(cfsfile, "package controllers")
		cfsfile.Close()

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

		fmt.Fprint(cgofile, clientfile)
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

		//setup the plugins
		RegisterDefaultPlugins(config.BuildPlugin)

		var kill = make(chan bool)
		go config.BuildPlugin.Activate(Plugins{Tag: "builder"}, config, kill)
		close(kill)
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
		fmt.Printf("--> Found app.yml and loading into config...\n")

		if err := config.Load(appfile); err != nil {
			fmt.Printf("--> --> ConfigError: 'app.yaml' -> %s\n", err)
			return
		}

		//setup the plugins
		RegisterDefaultPlugins(config.BuildPlugin)

		var kill = make(chan bool)

		//run the binary builder
		go config.BuildPlugin.Activate(Plugins{Tag: "watchBuildRun"}, config, kill)

		//run the js builder
		go config.BuildPlugin.Activate(Plugins{Tag: "jsWatchBuild"}, config, kill)

		//loadup the remaining plugins so they can activate themselves
		for _, mem := range config.Plugins {
			if mem.Tag == "watchBuildRun" || mem.Tag == "builder" {
				continue
			}
			go config.BuildPlugin.Activate(mem, config, kill)
		}

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGQUIT)
		signal.Notify(ch, syscall.SIGTERM)
		signal.Notify(ch, os.Interrupt)

		//setup a for loop and begin calling
		for {
			select {
			case <-ch:
				close(kill)
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
