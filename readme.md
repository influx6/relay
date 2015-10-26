#Relay
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/relay)
[![Travis](https://travis-ci.org/influx6/relay.svg?branch=master)](https://travis-ci.org/influx6/relay)

Relay is a simple microframework with very simple designs that provide you with the minimal tools you need to get things running.

# Download

     go get github.com/influx6/relay/...

# Install

     go install github.com/influx6/relay/...

# Changes

  - Added Build/Task plugin system:

      - builder(internal use only): This manages the building procedure when calling `relay build`

      - jsWatchBuild(internal use only): This manages the building of client files in ./client on changes within that directory,saving the output into ./static/js or as set in config

      - watchBuildRun(internal use only): This manages the building of entire codebase during use of `relay serve`

      - goFriday: This automatically converts your markdown files and saves them in order in a given directory, very useful for writing docs

      ```yaml

        # to automatically turn markdown files into go template files

        plugins:
          # using the custom markdown to gotemplate plugin
          markdown:
            tag: goFriday
            config:
              ext: ".tmpl"
              markdown: ./markdown
              templates: ./templates

      ```

      - commandWatch: This allows setting a series of commands to run on every changes in a directory path, to use just include in config as below:

      ```yaml

        # to execute and compile your less files

        plugins:
          less: #this is just a random tag you can assign
            tag: commandWatch #the plugin to use
            config: # a map of usable values, plugin defined
              path: "./static/less"
            args: # optional arguments but needed in this plugins use case
              - lessc ./static/less/main.less ./static/css/main.css
              - lessc ./static/less/svg.less ./static/css/svg.css
      ```



# Usage

   - Command Example:

     ```bash

          # once installation is done using 'go get'

          > relay
            λ relay
              relay provides a cli for relay projects

              Usage:
              relay [command]

              Available Commands:
              build       build the current relay project into a binary
              serve       serves up the project and watches for changes
              create      creates the relay project files and directory with the given name

              Flags:
              -h, --help[=false]: help for relay

              Use "relay [command] --help" for more information about a command.


          # to create a project directory just call the 'create' command giving the flag for the name of the project folder
          # and the --owner (i.e the name of your folder with the /src/github structure of go projects), this is used to
          # generate the package name and can be change accordingly in the "app.yaml" file

          > relay create --name wonderbat --owner influx6

            λ relay create --name wonderbat --owner influx6
              -> New relay Project: wonderbat, Owner: influx6 ...
              --> Creating 'wonderbat' project directory -> ./relay
              --> Creating 'bin' project directory
              --> Creating 'client' project directory
              --> Creating 'client/app' project directory
              --> Creating 'controllers' project directory
              --> Creating 'models' project directory
              --> Creating 'templates' project directory
              --> Creating 'static' project directory
              --> Creating 'vendor' project directory
              --> Creating project file: main.go
              --> Creating project file: controllers/controllers.go
              --> Creating project file: app.yml
              --> Creating project file: client/client.go
              --> Creating project file: client/app/app.go  


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

        ```



   - Where "app.yaml" contains =>


      ```yaml
                name: wonderbat
                addr: :4000
                env: dev

                hearbeat: 5m

                # directory and settings for static code
                static:
                  dir: ./static

                # directory to locate client gopherjs code
                client:
                    dir: ./client

                # change this to fit appropriately if using a different scheme
                package: github.com/influx6/wonderbat

      ```

# Example

  See [Relay Readme](https://github.com/influx6/relay/tree/master/relay/readme.md)

#License

  . [MIT](https://github.com/influx6/relay/blob/master/LICENSE)
