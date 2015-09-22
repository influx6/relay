#Relay
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/relay)

Relay is a simple microframework with very simple designs that provide you with the minimal tools you need to get things running.

# Download

     go get github.com/influx6/relay

# Install

     go install github.com/influx6/relay


#Features

  - Codecs: Relay builds on the ideas that response should be dynamic and customizable by providing encoders and decoders of various type that allow creating flexible and encapsulated message sub-protocols(not as complex as it sounds). It embodies the idea that messages should be easily retrieved or sent to and in appropriate format with a simple and clean

  - Middlewares: ingrained into the relay is the core strategy of every go http package but this is include as a optional tag on while staying pure at the router level. The middleware interface [FlatChain](https://github.com/influx6/relay/blob/master/middleware.go#L30) provides and empowers the bindings of controllers allowing extensive stacking of behaviours without overriding the ideal of simplicity

  - Controllers: this provide a basic encapsulation principle for a group of routine handlers and not in the sense of c of mvc. Controllers provide a simplified binding for handling connections (be it websocket or http) and embed routers to allow a more refined management of routers and subroutes

  - Engine: to simplify the startup and configuration of relay,a wrapper for the webserver is created to simplify the process, this will allow a more refined configuration system to be used with relay. Although still in flux, this system will allow more control of how the system works.

#Example

  ```yaml  
    #file: app.yaml

    addr: ":3000"

    folders:
      assets: ./app/assets
      models: ./app/models
      views: ./app/views

  ```

  ```go

    package app

    import (
    	"log"
    	"os"

    	"github.com/gorilla/websocket"
    	"github.com/influx6/relay"
    	"github.com/influx6/relay/engine"
    )

    conf := engine.NewConfig()

    //can be loaded from a file
  	if err := conf.Load("./app.yaml"); err != nil {
  		log.Printf("Error occured loading config: %s", err.Error())
  		return
  	}

  	app := engine.NewEngine(conf)

    //ServeDir serves a directory and ripps out the given path strip if supplied
  	app.ServeDir("/assets", app.Folders.Assets, "/assets/")

    //basic controller

    home := relay.NewController("/home")
    app.Bind(home)

    //the internal router allows the specification of
    //request methods to handle  for this route as a list of space seperated values
    app.Add("get head delete","/jails",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })

    //if its an empty string then all methods are allowed
    app.Add("","/updates",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })

    // Using Codecs
    //if its an empty string then all methods are allowed
    app.Add("","/updates",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })

    // Using the BindHTTP and BindSocket/UpgradeSocket provide custom handlers that use relays Context
    // and SocketWorker structs respectively allow a more refined and simple api call but also included
    //with context is an internal map to localize request data for easy access instead of a global map
    //of requests

    home.BindHTTP("get post put","/:id",func(req *relay.Context){
      //...handle the custom request object
      msg, err := BasicHTTPCodec.Deocde(req)
      //use the msg to get Params, Body or ParseForm/Form depending
      //on the content type of request (www-urlencode, multipart,body)
    }) // => returns a flatchain middleware

    //you can also use the pure go-handler approach
    home.Get("/updates",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })

    //for a more refined and simpler handler using the relay.Context struct which includes
    //a sync map for storing request data forthe lifetime of the requests
    home.BindHTTP("get post put","/:names",func(req *relay.Context,nx relay.NextHandler){
      //handle the custom request object
      nx(req)
    }) // => returns a FlatChain middleware

    ```

    - Using the relay codecs system:

    ```go

        home.BindHTTP("get post put","/:names",func(c *relay.Context,nx relay.NextHandler){
          //handle the custom request object
            json := JSONRender(200,map[string]string{"user":"john"}, true, true, true)

            //a special header encoder that uses the Context itself for some head writing
            BasicHeadEncoder.Encode(c, json.Head)

            //Encoders take in a io.Writer
            JSONEncoder.Encode(c.Res, json)
          nx(req)
        })

    ```

    - Using the middleware system in relay

    ```go

      // BindHTTP returns a middleware capable of providing stacking abilities
      home.BindHTTP("get post put","/:names",func(req *relay.Context,nx relay.NextHandler){
        //handle the custom request object
        nx(req)
      }).Chain(func(c *relay.Context,nx relay.NextHandler){
        //do something here...
        nx(c)
      })

      var Logger = NewFlatChain(func(c *relay.Context,next relay.NextHandler){
        log.Printf("Request: Method %s to %s",c.Req.Method,c.Req.URI.path)
        next(c)
      })

      home.Add("get post put","/:user",Logger.Handle)

      //continue chaining from that middleware
      Logger.Chain(func(c *relay.Context,nx relay.NextHandler){
        //do something here...
        nx(c)
      })
    ```

    - Using websockets:

    ```go

    //Binding for websocket connections exists and each controller provides
    //the BindSocket and BindSocketFor where each allows a more refined control on arguments.
    //Relay provides two strategies when dealing with websocket connections:

    //Strateg one: (handling socket messages directly)
    //With websocket is the SocketWorker which encapsulates
    //the gorilla.WebSocket object and create a infinite buffer
    //where messages are received until you begin to handle them
    //by receiving from the message channel
    //which returns a relay.WebsocketMessage type.
    //the handler is sent into a go-routine,so worry not of blockage ;)
    home.BindSocket("get post put","/socket",func(soc *relay.SocketWorker){

      for data := range soc.Messages {

        //the raw message provided by gorilla.webocket
        words := data.Message()

        if err != nil {
          continue
        }

        // if we use the default codec BasicSocketCodec,
        //then a []byte is returned but the codec used is up
        //to you,so this can be any thing your codec returns
        bu := words.([]byte)

        //using the gorilla.Conn wraped by relay.WebSocket directly
        err := data.Socket.WriteMessage(....)
      }


    },nil) // => returns a FlatChain middleware

    //Strategy two:
    //for a more chat like experience use for websocket, apart
    //from rolling out your own registration and broadcast units,
    //you can use the relay.SocketHub which takes each socket,registers it
    //and automatically receives messages from it and other registed sockets
    //and calls the supplied handler callback as below.
    //SocketHub provides a distribute function can exclude one supplied socket
    //from the message its sending to the others,to allow a reply approach

    //Has below lets create a central socket hub for the message and reply process of
    //the incoming sockets
    sockhub := relay.NewSocketHub(func(hub *relay.SocketHub, msg *relay.WebsocketMessage){
      //handle the message
      data,err := msg.Message()

      //distribute reply to others, but excluding the sender which can be
      //passed as the second argument
      hub.Distribute(func(other *relay.SocketWorker){

        //for more freedom you can write directly skipping the codec encoder
        other.Socket().WriteMessage(...)

      },msg.Worker)

    })

    home.BindSocket("get post put","/socket",hub.AddConnection,nil) // => returns a http middleware

    //BindSocketFor provides more control of what headers the socket uses,
    //the upgrade settings needed apart from the usual path,
    //request method and handler to use
    home.UpgradeSocket("","/socket",func(soc *relay.SocketWorker){
      //...
    },websocket.Upgrader{
    	ReadBufferSize:  1024,
    	WriteBufferSize: 1024,
    },http.Header(map[string][]string{
    	"Access-Control-Allow-Credentials": []string{"true"},
    	"Access-Control-Allow-Origin":      []string{"*"},
    })) //returns a FlatChains middleware


  	app.Serve()
  ```

#License

  . [MIT](https://github.com/influx6/relay/blob/master/LICENSE)
