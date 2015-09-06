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

    //the real route is /home/:id, using the internal request
    //encapsulation in relay caslled HTTPRequest, which wraps up the request
    //and response and uses the codecs for writing and reading data and provides a
    // .Message()(*Message,error) which parses the body or form data according to the
    //content type
    home.BindHTTP("get post put","/:id",func(req *relay.HTTPRequest){
      //...handle the custom request object
      msg, err := req.Message()
      //use the msg to get Params, Body or ParseForm/Form depending
      //content type of request
    },relay.BasicHTTPCodec) // => returns a http port

    //you can also use the pure go-handler approach
    home.Get("/updates",func(res http.ResponseWriter,req *http.Request,params relay.Collector){
      //handle the request and response
    })

    //the codec(last argument) argument can be nil which defaults to using the
    //BasicHTTPCodec as the internal http codec
    home.BindHTTP("get post put","/:names",func(req *relay.HTTPRequest){
      //handle the custom request object
    },nil) // => returns a http port

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

        //do something with the data and reply, since
        //relay.WebsocketMessage uses the internal or supplied codec,
        //the data can be anything you wish to return from the codec,
        //so a (interface{},error) is returned,so you can type asset
        words,err := data.Message()

        if err != nil {
          continue
        }

        // if we use the default codec BasicSocketCodec,
        //then a []byte is returned but the codec used is up
        //to you,so this can be any thing your codec returns
        bu := words.([]byte)

        //replies are written with the same message type
        //as the message it received
        //when using the 'WebsocketMessage.Write' method,
        data.Write([]byte("ok"))

        //or

        //using the gorilla.Conn wraped by relay.WebSocket directly
        err := data.Socket.WriteMessage(....)
      }


    },nil) // => returns a http port

    //Strategy two:
    //for a more chat like experience use for websocket, apart
    //from rolling out your own registration and broadcast units,
    //you can use the relay.SocketHub which takes each socket,registers
    //and automatically receives messages and calls a supplied callback
    //and provides a distribution function that excludes a supplied socket
    // create a central socket hub for the message and reply process
    sockhub := relay.NewSocketHub(func(hub *relay.SocketHub, msg *relay.WebsocketMessage){
      //handle the message
      data,err := msg.Message()

      //distribute reply to others, but excluding the sender which can be
      //passed as the second argument
      hub.Distribute(func(other *relay.SocketWorker){

        //for more freedom you can write directly skipping the codec encoder
        other.Socket().WriteMessage(...)

        //or

        //use the codec encoder but with more control of the type
        //of message we send,morphed and writing by the encoder.
        other.Write(gorilla.TextMessage,[]byte("yay! got it!"))

      },msg.Worker)

    })

    home.BindSocket("get post put","/socket",func(soc *relay.SocketWorker){
      hub.AddConnection(soc)
    },nil) // => returns a http port

    //BindSocketFor provides more control of what headers the socket uses,
    //the upgrade settings needed apart from the usual path,
    //request method and handler to use
    home.BindSocketFor("get post put","/socket",func(soc *relay.SocketWorker){
      //...
    },BasicSocketCodec,gorilla.Upgrader{
    	ReadBufferSize:  1024,
    	WriteBufferSize: 1024,
    },http.Header(map[string][]string{
    	"Access-Control-Allow-Credentials": []string{"true"},
    	"Access-Control-Allow-Origin":      []string{"*"},
    }))

  	app.Serve()
  ```

#License

  . [MIT]()
