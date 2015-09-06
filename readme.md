#Relay
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/relay)

Relay is a simple microframework with very simple designs that provide you with the minimal tools you need to get things running.

# Download

     go get github.com/influx6/relay

# Install

     go install github.com/influx6/relay


#Features

  - Codecs: Relay builds on the ideas that response should be dynamic and customizable,rather than creating a writer, response feed like approach, Relay uses codecs that take a response type or format and perform the operation of converting the response into the appropriate response object. This allows a custom message pattern to be created eg. binary only messages on top of http or websocket connections. Codecs have the pairs of encoders and decoders that provide the basis of transmission and reception.

  - Chainable Ports: Ports are just encapsulated, chain handlers of request which is a standard feature of most golang middleware based framework.

  - Controllers: this provide a basic encapsulation principle for a group of routine handlers and not in the sense of c of mvc. Controllers provide a simplified binding for handling connections (be it websocket or http) and embed routers to allow a more refined management of routers and subroutes

  - Engine: to simplify the startup and configuration of relay,a wrapper for the webserver is created to simplify the process, this will allow a more refined configuration system to be used with relay. Although still in flux, this system will allow more control of how the system works.

#Example

  ```go

    package app

    import (
    	"log"
    	"os"

    	"github.com/influx6/relay"
    	"github.com/influx6/relay/engine"
    )

  	app := engine.NewEngine()

    //ServeDir serves a directory and ripps out the given path strip if supplied
  	app.ServeDir("/assets", c.Folders.Assets, "/assets/")

    //basic controller

    home := relay.NewController("/home")
    app.Bind(home)

    //the internal router allows the specification of methods to handle as a list of space seperated values
    app.Add("get head delete","/jails",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })

    //if its an empty string then all methods are allowed
    app.Add("","/updates",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })

    //the real route is /home/:id
    //using the internal request encapsulation in relay which wraps up the request and response and uses the codecs for writing and reading data
    home.BindHTTP("get post put","/:id",func(req *relay.HTTPRequest){
      //handle the custom request object
    }) // => returns a http port

    //using the pure go-handler approach either
    home.Get("/updates",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })


  	app.Serve()
  ```

#License

  . [MIT]()
