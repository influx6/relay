#Relay
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/influx6/relay)
[![Travis](https://travis-ci.org/influx6/relay.svg?branch=master)](https://travis-ci.org/influx6/relay)

Relay is a simple microframework with very simple designs that provide you with the minimal tools you need to get things running.

# Download

     go get github.com/influx6/relay/...

# Install

     go install github.com/influx6/relay/...

# Example

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
    	"github.com/influx6/relay/relay"
    	"github.com/influx6/relay/engine"
    )

    conf := engine.NewConfig()

    //can be loaded from a file
  	if err := conf.Load("./app.yaml"); err != nil {
  		log.Printf("Error occured loading config: %s", err.Error())
  		return
  	}

  	app := engine.NewEngine(conf)

    app.Chain(relay.Logger(nil))

    //using the middleware router
		app.Rule("get head", "/favicon.ico", nil).Chain(relay.Redirect("/static/images/favicon.ico"))

    //the internal router allows the specification of
    //request methods to handle  for this route as a list of space seperated values
    app.Rule("get head delete","/jails",func(c *relay.Context,next relay.NextHandler){
      //handle the request and response
    })

    //if its an empty string then all methods are allowed
    app.Rule("","/updates",func(c *relay.Context,next relay.NextHandler){
      //handle the request and response
    })

    // Using Codecs
    //if its an empty string then all methods are allowed
    app.Rule("","/updates",func(c *relay.Context,next relay.NextHandler){
      //handle the request and response
    })


    ```

    - Using the relay codecs system:

    ```go

        app.Rule("get post put","/:names",func(c *relay.Context,nx relay.NextHandler){
          //handle the custom request object
            json := JSONRender(200,map[string]string{"user":"john"}, true, true, true)

            //a special header encoder that uses the Context itself for some head writing
            BasicHeadEncoder.Encode(c, json.Head)

            //Encoders take in a io.Writer
            JSONEncoder.Encode(c.Res, json)
          nx(req)
        })

    ```

    - Using websockets:

    ```go

    //use Link() to branch out into a new chain tree
    app.Rule("get post put","/socket").Link(relay.FlatSocket(nil,func(soc *relay.SocketWorker){

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


    },nil)) // => returns a FlatChain middleware

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

    app.Rule("get post put","/socket",nil).Link(relay.FlatSocket(nil,hub.AddConnection,nil))

  	app.Serve()
  ```

#License

  . [MIT](https://github.com/influx6/relay/blob/master/LICENSE)
